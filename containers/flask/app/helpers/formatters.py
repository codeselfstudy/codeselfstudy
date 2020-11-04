"""
Useful functions for formatting text.
"""

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
