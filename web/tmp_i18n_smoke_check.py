from playwright.sync_api import sync_playwright


BASE_URL = "http://127.0.0.1:4173"


def assert_contains(page, expected: str) -> None:
    body = page.locator("body").inner_text()
    if expected not in body:
        raise AssertionError(f"Expected to find: {expected}")


def check_public_route(page, path: str, expected_texts: list[str]) -> None:
    page.goto(f"{BASE_URL}{path}", wait_until="networkidle")
    for text in expected_texts:
        assert_contains(page, text)


with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()

    check_public_route(
        page,
        "/es/",
        [
            "Correo: support@dotblue.ai",
            "Dotblue Tech Co., Ltd.",
        ],
    )

    check_public_route(
        page,
        "/zh-CN/",
        [
            "邮箱：support@dotblue.ai",
            "Dotblue Tech Co., Ltd.",
        ],
    )

    browser.close()
