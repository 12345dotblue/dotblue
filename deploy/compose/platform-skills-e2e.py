from __future__ import annotations

import argparse
import json
import re
import urllib.error
import urllib.request
from pathlib import Path

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from playwright.sync_api import sync_playwright


REPO_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_ENV_FILE = Path(__file__).with_name(".env")
DEFAULT_CONFIG_FILE = Path(__file__).parent / ".generated" / "dotblue" / "config.yaml"
DEFAULT_ARTIFACT_DIR = REPO_ROOT / ".tmp" / "webapp-testing" / "platform-skills-e2e"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run the platform skill reference-cycle UI E2E flow against the compose deployment.",
    )
    parser.add_argument("--env-file", type=Path, default=DEFAULT_ENV_FILE)
    parser.add_argument("--config-file", type=Path, default=DEFAULT_CONFIG_FILE)
    parser.add_argument("--artifact-dir", type=Path, default=DEFAULT_ARTIFACT_DIR)
    parser.add_argument("--source-skill-code", default="partner.petstore")
    parser.add_argument("--source-version", default="")
    parser.add_argument("--target-skill-code", default="demo.skill")
    parser.add_argument("--target-version", default="")
    parser.add_argument("--language", default="zh-CN")
    parser.add_argument("--expected-save-message", default="引用关系已更新。")
    parser.add_argument("--expected-publish-error", default="引用关系存在环，无法发布。")
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
    if not config_path.exists():
        raise FileNotFoundError(f"Missing generated config: {config_path}")

    text = config_path.read_text(encoding="utf-8")
    client_id_match = re.search(r'clientId:\s*"([^"]+)"', text)
    client_secret_match = re.search(r'clientSecret:\s*"([^"]+)"', text)
    if not client_id_match or not client_secret_match:
        raise RuntimeError(f"Failed to read oauth client from {config_path}")
    return client_id_match.group(1), client_secret_match.group(1)


def build_settings(args: argparse.Namespace) -> dict[str, str | Path]:
    env = load_env_file(args.env_file)
    client_id, client_secret = read_oauth_client(args.config_file)

    return {
        "base_url": env["DOTBLUE_PUBLIC_URL"].rstrip("/"),
        "backend_url": env["DOTBLUE_BACKEND_PUBLIC_URL"].rstrip("/"),
        "casdoor_url": env["CASDOOR_PUBLIC_URL"].rstrip("/"),
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


def visible_texts(locator, limit: int = 10) -> list[str]:
    texts: list[str] = []
    count = min(locator.count(), limit)
    for index in range(count):
        try:
            texts.append(locator.nth(index).inner_text().strip())
        except Exception as err:  # debug helper
            texts.append(f"<error:{err}>")
    return texts


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


def get_json(url: str, token: str):
    return request_json(url, headers={"Authorization": f"Bearer {token}"})


def post_json(url: str, payload: dict | list, token: str):
    return request_json(
        url,
        method="POST",
        payload=payload,
        headers={"Authorization": f"Bearer {token}"},
    )


def resolve_skill(settings: dict[str, str | Path], token: str, code: str) -> dict:
    skills = get_json(f"{settings['backend_url']}/api/admin/skills?view=governance", token)
    for skill in skills:
        if skill.get("code") == code:
            detail = get_json(
                f"{settings['backend_url']}/api/admin/skills/{skill['id']}",
                token,
            )
            return detail
    available_codes = ", ".join(sorted(item.get("code", "") for item in skills))
    raise RuntimeError(f"Skill code not found: {code}. Available: {available_codes}")


def select_version(detail: dict, requested_version: str, preferred_statuses: tuple[str, ...]) -> dict:
    versions = detail.get("versions") or []
    if not versions:
        raise RuntimeError(f"Skill has no versions: {detail['skill']['code']}")

    if requested_version:
        for version in versions:
            if version.get("version") == requested_version:
                return version
        raise RuntimeError(
            f"Requested version not found: {detail['skill']['code']}@{requested_version}",
        )

    for status in preferred_statuses:
        for version in versions:
            if version.get("releaseStatus") == status:
                return version
    return versions[0]


def select_target_version_id(detail: dict, requested_version: str) -> str:
    skill = detail["skill"]
    if requested_version:
        return select_version(detail, requested_version, ("published",))["id"]

    if skill.get("latestPublishedVersionId"):
        return skill["latestPublishedVersionId"]

    for version in detail.get("versions") or []:
        if version.get("releaseStatus") == "published":
            return version["id"]

    raise RuntimeError(f"Target skill has no published version: {skill['code']}")


def reset_references(settings: dict[str, str | Path], token: str, skill_id: str, version_id: str) -> None:
    # Always clean up by API so repeated runs start from a known graph state.
    post_json(
        f"{settings['backend_url']}/api/admin/skills/{skill_id}/references",
        {"skillVersionId": version_id, "references": []},
        token,
    )


def wait_for_message(page, text: str, timeout_ms: int = 8000) -> None:
    page.get_by_text(text, exact=False).wait_for(timeout=timeout_ms)


def main() -> int:
    args = parse_args()
    settings = build_settings(args)
    artifact_dir = Path(settings["artifact_dir"])
    ensure_artifact_dir(artifact_dir)

    token = get_admin_token(settings)
    source_detail = resolve_skill(settings, token, args.source_skill_code)
    target_detail = resolve_skill(settings, token, args.target_skill_code)
    source_version = select_version(source_detail, args.source_version, ("draft", "reviewing"))
    target_version_id = select_target_version_id(target_detail, args.target_version)

    base_url = str(settings["base_url"])
    language = str(settings["language"])
    login_url = f"{base_url}/{language}/login"
    skills_url = f"{base_url}/{language}/admin/platform/skills"
    publish_url = (
        f"{settings['backend_url']}/api/admin/skills/{source_detail['skill']['id']}/publish"
    )
    references_url = (
        f"{settings['backend_url']}/api/admin/skills/{source_detail['skill']['id']}/references"
    )

    try:
        with sync_playwright() as playwright:
            browser = playwright.chromium.launch(headless=True)
            context = browser.new_context(viewport={"width": 1440, "height": 1400})
            page = context.new_page()

            print(f"[info] source skill: {source_detail['skill']['code']}@{source_version['version']}")
            print(f"[info] target skill: {target_detail['skill']['code']} -> {target_version_id}")

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

            try:
                page.wait_for_url("**/callback?**", timeout=15000)
            except PlaywrightTimeoutError:
                print(f"[warn] callback url not observed, current url: {page.url}")

            try:
                page.wait_for_url("**/dashboard", timeout=20000)
            except PlaywrightTimeoutError:
                print(f"[warn] dashboard url not observed, current url: {page.url}")
            page.wait_for_load_state("networkidle")
            print(f"[info] after login url: {page.url}")
            save_debug(page, artifact_dir, "02-dashboard")

            print(f"[step] open skills page: {skills_url}")
            page.goto(skills_url, wait_until="domcontentloaded")
            page.wait_for_load_state("networkidle")
            print(f"[info] skills page url: {page.url}")
            print(f"[info] visible buttons: {visible_texts(page.locator('button'))}")
            save_debug(page, artifact_dir, "03-skills")

            source_row = page.locator("tbody tr", has_text=source_detail["skill"]["code"]).first
            source_row.wait_for(timeout=15000)

            print("[step] open source skill detail")
            with page.expect_response(
                lambda resp: resp.url.endswith(f"/api/admin/skills/{source_detail['skill']['id']}")
                and resp.request.method == "GET",
                timeout=15000,
            ) as detail_response:
                source_row.get_by_role("button", name="查看详情").click()
            print(f"[info] skill detail status: {detail_response.value.status} -> {detail_response.value.url}")
            page.wait_for_load_state("networkidle")
            save_debug(page, artifact_dir, "04-skill-detail")

            version_row = page.locator(".ant-drawer .ant-table-tbody tr", has_text=source_version["version"]).first
            version_row.wait_for(timeout=15000)

            print("[step] open reference editor")
            with page.expect_response(
                lambda resp: resp.url.endswith(
                    f"/api/admin/skills/{source_detail['skill']['id']}/versions/{source_version['id']}/references"
                ) and resp.request.method == "GET",
                timeout=15000,
            ) as references_response:
                version_row.get_by_role("button", name="编辑引用").click()
            print(
                f"[info] version references status: {references_response.value.status} -> {references_response.value.url}"
            )
            page.wait_for_load_state("networkidle")
            save_debug(page, artifact_dir, "05-reference-editor")

            textarea = page.locator(".ant-modal textarea").first
            textarea.wait_for(timeout=15000)
            print(f"[info] reference textarea value length before save: {len(textarea.input_value())}")

            # The target published version already references the source in the seeded data,
            # so saving this reverse edge must trigger publish-time cycle detection.
            cycle_reference = json.dumps(
                [
                    {
                        "toSkillVersionId": target_version_id,
                        "invokeMode": "sync",
                        "conditionExpr": "",
                        "contextPassthrough": True,
                        "resultPassthrough": False,
                        "sortOrder": 1,
                    }
                ],
                ensure_ascii=False,
                indent=2,
            )
            textarea.fill(cycle_reference)

            print("[step] save cycle reference from UI")
            with page.expect_response(
                lambda resp: resp.url == references_url and resp.request.method == "POST",
                timeout=15000,
            ) as save_response:
                page.locator(".ant-modal .ant-btn-primary").click()
            print(f"[info] reference save status: {save_response.value.status} -> {save_response.value.url}")
            wait_for_message(page, args.expected_save_message)
            save_debug(page, artifact_dir, "06-reference-saved")

            print("[step] publish source version to trigger cycle prevention")
            with page.expect_response(
                lambda resp: resp.url == publish_url and resp.request.method == "POST",
                timeout=15000,
            ) as publish_response:
                version_row.get_by_role("button", name="发 布").click()
            print(f"[info] publish status: {publish_response.value.status} -> {publish_response.value.url}")
            publish_body = publish_response.value.text()
            print(f"[info] publish response body: {publish_body}")

            if publish_response.value.status != 400:
                raise RuntimeError(f"Expected publish to fail with 400, got {publish_response.value.status}")
            if "skill reference cycle detected" not in publish_body:
                raise RuntimeError(f"Unexpected publish response body: {publish_body}")

            wait_for_message(page, args.expected_publish_error)
            save_debug(page, artifact_dir, "07-publish-cycle-error")
    finally:
        print("[step] cleanup source references")
        reset_references(
            settings,
            token,
            source_detail["skill"]["id"],
            source_version["id"],
        )
        print(f"[done] artifacts saved in: {artifact_dir}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
