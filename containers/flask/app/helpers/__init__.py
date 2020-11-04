"""Helper functions"""
# TODO: move these functions to other modules.


def safe_list_get(lst, idx, default):
    """A helper function to safely get the first element of a list."""
    if not lst:
        return default

    try:
        return lst[idx]
    except IndexError:
        return default


def format_codewars_puzzle_message(puzzle):
    print("helper got puzzle", puzzle)
    if not puzzle:
        return None

    languages = ", ".join(puzzle["languages"])
    lines = [
        f"*{puzzle['name']}* ({puzzle['kyu']} kyu)",
        f"{puzzle['url']}"
        f"\n",
        f"> *available in:* {languages}",
        f"> *category:* {puzzle['category']}"
    ]
    return "\n".join(lines)
