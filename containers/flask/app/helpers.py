"""Helper functions"""


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
    name, kyu, languages, category, url = puzzle
    languages_string = ",".join("languages")
    lines = [
        f"*{name}",
        f"- kyu: {kyu}",
        f"- languages: {languages_string}",
        f"- category: {category}",
        f"- url: {url}"
    ]
    return "\n".join(lines)
