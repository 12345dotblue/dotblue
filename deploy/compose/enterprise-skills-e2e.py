from __future__ import annotations

import argparse
import json
import re
import time
import urllib.request
from pathlib import Path

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from playwright.sync_api import sync_playwright


REPO_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_ENV_FILE = Path(__file__).with_name(".env")
DEFAULT_CONFIG_FILE = Path(__file__).parent / ".generated" / "dotblue" / "config.yaml"
DEFAULT_ARTIFACT_DIR = REPO_ROOT / ".tmp" / "webapp-testing" / "enterprise-skills-e2e"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run the enterprise skill create/version/review/publish/catalog UI E2E flow.",
    )
    parser.add_argument("--env-file", type=Path, default=DEFAULT_ENV_FILE)
    parser.add_argument("--config-file", type=Path, default=DEFAULT_CONFIG_FILE)
    parser.add_argument("--artifact-dir", type=Path, default=DEFAULT_ARTIFACT_DIR)
    parser.add_argument("--language", default="zh-CN")
    parser.add_argument("--base-url", default="http://localhost:9000")
    parser.add_argument("--backend-url", default="")
    parser.add_argument("--casdoor-url", default="")
    return parser.parse_args()


def load_env_file(path: Path) -> dict[str, str]:
    if not path.exists():
        raise FileNotFoundError(f"Missing env file: {path}")

    env: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        env[key.strip()] = value.strip()
    return env


def read_oauth_client(config_path: Path) -> tuple[str, str]:
    text = config_path.read_text(encoding="utf-8")
    client_id_match = re.search(r'clientId:\s*"([^"]+)"', text)
    client_secret_match = re.search(r'clientSecret:\s*"([^"]+)"', text)
    if not client_id_match or not client_secret_match:
        raise RuntimeError(f"Failed to read oauth client from {config_path}")
    return client_id_match.group(1), client_secret_match.group(1)


def build_settings(args: argparse.Namespace) -> dict[str, str | Path]:
    env = load_env_file(args.env_file)
    client_id, client_secret = read_oauth_client(args.config_file)
    backend_url = args.backend_url or env["DOTBLUE_BACKEND_PUBLIC_URL"]
    casdoor_url = args.casdoor_url or env["CASDOOR_PUBLIC_URL"]
    return {
        "base_url": args.base_url.rstrip("/"),
        "backend_url": backend_url.rstrip("/"),
        "casdoor_url": casdoor_url.rstrip("/"),
        "username": env["DOTBLUE_ADMIN_USERNAME"],
        "password": env["DOTBLUE_ADMIN_PASSWORD"],
        "client_id": client_id,
        "client_secret": client_secret,
        "language": args.language,
        "artifact_dir": args.artifact_dir,
    }


def ensure_artifact_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def save_debug(page, artifact_dir: Path, name: str) -> None:
    page.screenshot(path=str(artifact_dir / f"{name}.png"), full_page=True)
    (artifact_dir / f"{name}.html").write_text(page.content(), encoding="utf-8")


def request_json(
    url: str,
    method: str = "GET",
    payload: dict | list | None = None,
    headers: dict[str, str] | None = None,
):
    data = None
    request_headers = {"Accept": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        request_headers["Content-Type"] = "application/json"
    for key, value in (headers or {}).items():
        request_headers[key] = value

    request = urllib.request.Request(url, data=data, method=method)
    for key, value in request_headers.items():
        request.add_header(key, value)

    with urllib.request.urlopen(request) as response:
        body = response.read().decode("utf-8")
        return json.loads(body) if body else {}


def get_admin_token(settings: dict[str, str | Path]) -> str:
    token_response = request_json(
        f"{settings['casdoor_url']}/api/login/oauth/access_token",
        method="POST",
        payload={
            "grant_type": "password",
            "client_id": settings["client_id"],
            "client_secret": settings["client_secret"],
            "username": settings["username"],
            "password": settings["password"],
            "scope": "openid profile email",
        },
    )
    return token_response["access_token"]


def wait_for_message(page, text: str, timeout_ms: int = 8000) -> None:
    page.get_by_text(text, exact=False).wait_for(timeout=timeout_ms)


def wait_for_dialog(page, name_pattern, timeout_ms: int = 15000):
    dialog = page.get_by_role("dialog", name=name_pattern).last
    dialog.wait_for(state="visible", timeout=timeout_ms)
    return dialog


def main() -> int:
    args = parse_args()
    settings = build_settings(args)
    artifact_dir = Path(settings["artifact_dir"])
    ensure_artifact_dir(artifact_dir)

    timestamp = int(time.time())
    skill_code = f"enterprise.e2e.{timestamp}"
    skill_name = f"企业联调 Skill {timestamp}"
    skill_version = "1.0.0"
    login_url = f"{settings['base_url']}/{settings['language']}/login"
    enterprise_skills_url = f"{settings['base_url']}/{settings['language']}/admin/enterprise?tab=skills"
    token = get_admin_token(settings)

    try:
        with sync_playwright() as playwright:
            browser = playwright.chromium.launch(headless=True)
            context = browser.new_context(viewport={"width": 1440, "height": 1400})
            page = context.new_page()
            page.on(
                "response",
                lambda response: print(f"[net] {response.request.method} {response.status} {response.url}")
                if "/api/admin/skills" in response.url
                else None,
            )
            page.on(
                "console",
                lambda message: print(f"[console:{message.type}] {message.text}")
                if message.type in {"error", "warning"}
                else None,
            )

            print(f"[info] skill code: {skill_code}")
            print(f"[step] open login: {login_url}")
            page.goto(login_url, wait_until="domcontentloaded")
            page.wait_for_load_state("networkidle")
            save_debug(page, artifact_dir, "01-login")

            page.get_by_role("textbox", name="username, Email or phone").fill(str(settings["username"]))
            page.get_by_role("textbox", name="Password").fill(str(settings["password"]))

            print("[step] submit login")
            with page.expect_response(
                lambda resp: "/api/login?" in resp.url and resp.request.method == "POST",
                timeout=15000,
            ) as login_response:
                page.get_by_role("button", name="Sign In").click()
            print(f"[info] casdoor login status: {login_response.value.status}")

            page.wait_for_url("**/callback?**", timeout=15000)
            page.wait_for_url("**/dashboard", timeout=20000)
            page.wait_for_load_state("networkidle")
            save_debug(page, artifact_dir, "01-dashboard")

            print(f"[step] open enterprise skills: {enterprise_skills_url}")
            page.goto(enterprise_skills_url, wait_until="domcontentloaded")
            page.wait_for_load_state("networkidle")
            page.get_by_role("heading", name="企业自有 Skill").wait_for(timeout=15000)
            enterprise_id = page.evaluate("() => window.localStorage.getItem('dotblue_current_enterprise_id') || ''")
            if not enterprise_id:
                raise RuntimeError("Current enterprise id missing in browser storage")
            save_debug(page, artifact_dir, "02-enterprise-skills")

            print("[step] create enterprise skill")
            page.get_by_role("button", name="新建企业 Skill").last.click()
            modal = wait_for_dialog(page, re.compile(r"新建企业 Skill"))
            code_input = modal.locator("input#code")
            code_input.wait_for(state="visible", timeout=15000)
            save_debug(page, artifact_dir, "03-create-modal")
            code_input.fill(skill_code)
            modal.locator("input#name").fill(skill_name)
            modal.locator("textarea#description").fill("企业 Skill 真实联调创建")
            modal.locator(".ant-modal-footer .ant-btn-primary").click(force=True)
            page.wait_for_timeout(1000)
            validation_errors = modal.locator(".ant-form-item-explain-error").all_text_contents()
            if validation_errors:
                raise RuntimeError(f"Create skill validation blocked submit: {validation_errors}")
            page.get_by_text(skill_code, exact=False).wait_for(timeout=15000)
            save_debug(page, artifact_dir, "03-created")

            skill_detail = request_json(
                f"{settings['backend_url']}/api/admin/skills?view=governance",
                headers={
                    "Authorization": f"Bearer {token}",
                    "X-Enterprise-ID": str(enterprise_id),
                },
            )
            created = next((item for item in skill_detail if item.get("code") == skill_code), None)
            if not created:
                raise RuntimeError(f"Created skill not found in governance list: {skill_code}")
            skill_id = created["id"]

            row = page.locator("tbody tr", has_text=skill_code).first
            row.wait_for(timeout=15000)

            print("[step] open skill detail")
            with page.expect_response(
                lambda resp: resp.url.endswith(f"/api/admin/skills/{skill_id}") and resp.request.method == "GET",
                timeout=15000,
            ) as detail_response:
                row.get_by_role("button", name="查看详情").click()
            print(f"[info] detail status: {detail_response.value.status}")
            detail_modal = wait_for_dialog(page, re.compile(re.escape(skill_name)))
            detail_modal.get_by_role("button", name="新建版本").wait_for(timeout=15000)
            save_debug(page, artifact_dir, "04-detail")

            print("[step] create version")
            detail_modal.get_by_role("button", name="新建版本").click()
            modal = wait_for_dialog(page, re.compile(r"新建版本"))
            version_input = modal.locator("input#version")
            version_input.wait_for(state="visible", timeout=15000)
            version_input.fill(skill_version)
            modal.locator("textarea#changeLog").fill("企业 Skill 联调版本")
            save_debug(page, artifact_dir, "05-version-modal")
            with page.expect_response(
                lambda resp: resp.url.endswith(f"/api/admin/skills/{skill_id}/versions") and resp.request.method == "POST",
                timeout=15000,
            ) as version_response:
                modal.locator(".ant-modal-footer .ant-btn-primary").click(force=True)
            print(f"[info] create version status: {version_response.value.status}")
            if version_response.value.status >= 300:
                raise RuntimeError(f"Create version failed: {version_response.value.status} -> {version_response.value.text()}")
            wait_for_message(page, "Skill 版本创建成功")

            version_row = page.locator(".ant-modal-root .ant-table-tbody tr", has_text=skill_version).first
            try:
                version_row.wait_for(timeout=3000)
            except PlaywrightTimeoutError:
                version_row = page.locator(".ant-table-tbody tr", has_text=skill_version).first
                version_row.wait_for(timeout=15000)
            save_debug(page, artifact_dir, "05-version-created")

            print("[step] submit review")
            with page.expect_response(
                lambda resp: resp.url.endswith(f"/api/admin/skills/{skill_id}/submit-review") and resp.request.method == "POST",
                timeout=15000,
            ) as review_response:
                version_row.get_by_role("button", name="提交审核").click()
            print(f"[info] submit review status: {review_response.value.status}")
            if review_response.value.status >= 300:
                raise RuntimeError(f"Submit review failed: {review_response.value.status} -> {review_response.value.text()}")
            wait_for_message(page, "Skill 版本已提交审核")
            save_debug(page, artifact_dir, "06-review-submitted")

            print("[step] publish version")
            version_row = page.locator(".ant-table-tbody tr", has_text=skill_version).first
            version_row.wait_for(timeout=15000)
            with page.expect_response(
                lambda resp: resp.url.endswith(f"/api/admin/skills/{skill_id}/publish") and resp.request.method == "POST",
                timeout=15000,
            ) as publish_response:
                publish_button = version_row.locator("button").last
                publish_button.wait_for(timeout=15000)
                publish_button.click()
            print(f"[info] publish status: {publish_response.value.status}")
            if publish_response.value.status >= 300:
                raise RuntimeError(f"Publish failed: {publish_response.value.status} -> {publish_response.value.text()}")
            wait_for_message(page, "Skill 版本发布成功")
            save_debug(page, artifact_dir, "07-published")

            detail_modal.locator(".ant-modal-close").click()
            detail_modal.wait_for(state="hidden", timeout=15000)

            print("[step] switch to catalog view")
            page.get_by_role("button", name=re.compile(r"目录视图")).click()
            page.wait_for_load_state("networkidle")
            page.get_by_role("heading", name="可消费 Skill 目录").wait_for(timeout=15000)
            catalog_row = page.locator("tbody tr", has_text=skill_code).first
            catalog_row.wait_for(timeout=15000)
            save_debug(page, artifact_dir, "08-catalog")
            row_text = catalog_row.text_content() or ""
            disable_count = catalog_row.get_by_role("button", name="停用").count()
            print(f"[info] catalog row text: {row_text}")
            print(f"[info] catalog disable button count: {disable_count}")
            if disable_count == 0 and "已启用" not in row_text and "启用中" not in row_text:
                raise RuntimeError("Expected enterprise-owned published skill to appear as enabled in catalog view")

            print("[done] enterprise skill UI E2E passed")
            browser.close()
    finally:
        print(f"[done] artifacts saved in: {artifact_dir}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
