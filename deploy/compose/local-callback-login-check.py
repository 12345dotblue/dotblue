from __future__ import annotations

import argparse
from pathlib import Path

from playwright.sync_api import TimeoutError, sync_playwright


def save_debug(page, artifact_dir: Path, name: str) -> None:
    artifact_dir.mkdir(parents=True, exist_ok=True)
    page.screenshot(path=str(artifact_dir / f"{name}.png"), full_page=True)
    (artifact_dir / f"{name}.html").write_text(page.content(), encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate local Casdoor callback login and enterprise skills page")
    parser.add_argument("--base-url", default="http://localhost:9000")
    parser.add_argument("--language", default="zh-CN")
    parser.add_argument("--username", required=True)
    parser.add_argument("--password", required=True)
    parser.add_argument(
        "--artifact-dir",
        default=str(Path("c:/Users/kongz/work/dotblue/.tmp/webapp-testing/local-callback-login-check")),
    )
    args = parser.parse_args()

    artifact_dir = Path(args.artifact_dir)
    login_url = f"{args.base_url}/{args.language}/login"
    enterprise_skills_url = f"{args.base_url}/{args.language}/admin/enterprise?tab=skills"

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1440, "height": 1400})
        page = context.new_page()

        try:
            print(f"[step] open login: {login_url}")
            page.goto(login_url, wait_until="domcontentloaded")
            page.wait_for_load_state("networkidle")
            save_debug(page, artifact_dir, "01-login")

            page.get_by_role("textbox", name="username, Email or phone").fill(args.username)
            page.get_by_role("textbox", name="Password").fill(args.password)

            print("[step] submit login")
            with page.expect_response(
                lambda resp: "/api/login?" in resp.url and resp.request.method == "POST",
                timeout=15000,
            ) as login_response:
                page.get_by_role("button", name="Sign In").click()
            print(f"[info] casdoor login status: {login_response.value.status}")

            page.wait_for_url("**/callback?**", timeout=15000)
            print(f"[info] callback url: {page.url}")
            page.wait_for_url("**/dashboard", timeout=20000)
            page.wait_for_load_state("networkidle")
            print(f"[info] dashboard url: {page.url}")
            save_debug(page, artifact_dir, "02-dashboard")

            print(f"[step] open enterprise skills: {enterprise_skills_url}")
            page.goto(enterprise_skills_url, wait_until="domcontentloaded")
            page.wait_for_load_state("networkidle")
            page.get_by_role("heading", name="企业自有 Skill").wait_for(timeout=15000)
            save_debug(page, artifact_dir, "03-enterprise-skills")
            print(f"[info] enterprise skills url: {page.url}")
            return 0
        except TimeoutError as exc:
            save_debug(page, artifact_dir, "99-timeout")
            print(f"[error] timeout: {exc}")
            return 1
        except Exception as exc:  # pragma: no cover - debugging path
            save_debug(page, artifact_dir, "99-error")
            print(f"[error] unexpected: {exc}")
            return 1
        finally:
            context.close()
            browser.close()


if __name__ == "__main__":
    raise SystemExit(main())
