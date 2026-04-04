from app.prompts import SYSTEM_PROMPT, build_duplicate_nudge, build_user_prompt


def test_system_prompt_mentions_tools():
    assert "search_code" in SYSTEM_PROMPT
    assert "read_file" in SYSTEM_PROMPT
    assert "grep" in SYSTEM_PROMPT
    assert "run_tests" in SYSTEM_PROMPT


def test_system_prompt_mentions_diagnosis():
    assert "diagnosis" in SYSTEM_PROMPT.lower() or "diagnose" in SYSTEM_PROMPT.lower()


def test_build_user_prompt_with_error():
    prompt = build_user_prompt(
        description="upload returns 500",
        error_output="Traceback: ValueError in parser.py",
    )
    assert "upload returns 500" in prompt
    assert "Traceback" in prompt


def test_build_user_prompt_without_error():
    prompt = build_user_prompt(
        description="upload returns 500",
        error_output=None,
    )
    assert "upload returns 500" in prompt
    lower = prompt.lower()
    assert "error output" not in lower or "none" in lower or "no error" in lower


def test_build_duplicate_nudge():
    nudge = build_duplicate_nudge("search_code", '{"query": "foo"}')
    assert "search_code" in nudge
    assert "different" in nudge.lower() or "another" in nudge.lower()
