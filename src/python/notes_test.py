from time import sleep

import pytest
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

INDEX_URL = "http://localhost:8040"


@pytest.fixture()
def browser():
    options = Options()
    options.add_argument("--headless")
    options.add_argument("window-size=1920,1080")
    browser = webdriver.Chrome(options=options)
    yield browser
    browser.quit()


def upload_file(browser, name, path):
    content = open(path).read()
    browser.execute_script("""
    var dropzone_instance = Dropzone.forElement("#userInputForm")
    var new_file = new File(['%s'], '%s', {type: 'application/octet-binary'})
    dropzone_instance.addFile(new_file)
    """ % (content, name))


def assert_no_javascript_errors(browser):
    assert len(browser.find_elements_by_css_selector(".js-error")) == 0, (
        "Javascript errors found. See browser console for more details."
    )


def test_index_should_redirect_to_edit(browser):
    browser.get(INDEX_URL)
    assert "/edit" in browser.current_url


def test_should_edit(browser):
    browser.get(INDEX_URL)
    assert_no_javascript_errors(browser)

    # Type a new note
    user_input = browser.find_element_by_id("userInput")
    user_input.clear()
    user_input.send_keys(
        "# Hello world\n\n"
        "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec feugiat "
        "euismod nibh, ac scelerisque erat laoreet quis. Nullam dignissim varius "
        "enim. Aenean interdum et elit eu gravida. Cras eleifend eget tortor sit amet "
        "tincidunt. Praesent eu interdum turpis. Nullam et massa massa. Maecenas "
        "maximus turpis id egestas rhoncus. Morbi eget bibendum leo. "
    )
    sleep(1)
    assert_no_javascript_errors(browser)

    # Should render the preview
    h1 = browser.find_element_by_css_selector("h1")
    assert h1.text == "Hello world"

    # Refresh should not delete the content
    browser.refresh()
    assert "Lorem ipsum" in browser.page_source

    # Click view and verify content
    browser.find_element_by_link_text("View").click()
    assert_no_javascript_errors(browser)
    assert "/view" in browser.current_url
    h1 = browser.find_element_by_css_selector("h1")
    assert h1.text == "Hello world"
    assert "Lorem ipsum" in browser.page_source
    assert "Hello world" in browser.title

    # Click publish and verify content
    browser.find_element_by_link_text("Publish").click()
    assert_no_javascript_errors(browser)
    h1 = browser.find_element_by_css_selector("h1")
    assert h1.text == "Hello world"
    assert "Hello world" in browser.title


def test_upload(browser):
    browser.get(INDEX_URL)
    assert_no_javascript_errors(browser)

    # Upload a file
    upload_file(browser, "readme.txt", "src/python/assets/readme.txt")
    sleep(1)
    assert_no_javascript_errors(browser)

    # Should create a link
    browser.find_element_by_link_text("readme.txt")


def test_code_highlight(browser):
    browser.get(INDEX_URL)
    assert_no_javascript_errors(browser)

    # Type some source code
    user_input = browser.find_element_by_id("userInput")
    user_input.clear()
    user_input.send_keys(
        "```python\n"
        "def hello_world():\n"
        "    print('Hello')\n"
        "```\n"
    )
    sleep(1)
    assert_no_javascript_errors(browser)

    # Should create highlighted elements
    browser.find_element_by_css_selector(".hljs-title")
