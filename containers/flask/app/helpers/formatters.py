"""
Useful functions for formatting text.
"""
import os
# from textwrap import dedent

DISCOURSE_PUZZLES_CATEGORY = os.getenv("DISCOURSE_PUZZLES_CATEGORY")
print("puzzles category", DISCOURSE_PUZZLES_CATEGORY)

def format_codewars_puzzle_message(puzzle):
    """This formats a codewars puzzle for Slack."""
    print("helper got puzzle", puzzle)
    if not puzzle:
        return None

    languages = ", ".join(puzzle["languages"])
    # TODO: this hack could be cleaned up with the `dedent` function
    lines = [
        f"*{puzzle['name']}* ({puzzle['kyu']} kyu)",
        f"{puzzle['url']}"
        f"\n",
        f"> *available in:* {languages}",
        f"> *category:* {puzzle['category']}"
    ]
    return "\n".join(lines)


def format_codewars_puzzle_for_discourse(puzzle):
    """Format a codewars puzzle to post as a forum post in Discourse."""
    print("helper got puzzle for discourse", puzzle)
    if not puzzle:
        return None

    title = f"Puzzle: {puzzle['name']} ({puzzle['category']})"
    languages = ", ".join(puzzle["languages"])
    tags = ", ".join(puzzle['tags'])

    description = puzzle.get("description", None)
    if description:
        description = description.replace(r"```", "\n```\n")
    # `dedent` wasn't working for me with format strings, so I'm removing the indents manually
    lines = f"""**{puzzle["name"]}** is a coding puzzle that people can be attempted in the following languages:
        {languages}

        - **Difficulty:** {puzzle.get("kyu", "unknown")} kyu
        - **Stars:** {puzzle.get("stars", "unknown")}
        - **Votes:** {puzzle.get("votes", "unknown")}
        - **Tags:** {tags}

        ## Description

        {description}

        ## Link

        {puzzle["url"]}

        ## Notes

        This puzzle was posted by a Slackbot. If you want to help work on the app, send a message to @Josh.

        You can discuss the problem and solutions in the comments below. Use the "hide details" feature in the editor to avoid spoilers.
        """.split("\n")

    cleaned_lines = [line.strip() for line in lines]
    raw = "\n".join(cleaned_lines)

    return {
        "title": title,
        "raw": raw,
        "category": DISCOURSE_PUZZLES_CATEGORY
    }


def format_slack_error_message():
    """This is the message that gets sent back to Slack when there is an error."""
    return dedent("""
        Your query wasn't formatted correctly or there was another error. Try typing something like this:

        ```/puzzle codewars js python elixir 5kyu 100votes```

        Parameters can look like this:
        - *source:* `codewars` (required)
        - *languages:* `python js fortran c` (optional)
        - *difficulty:* `5kyu` (optional)
        - *minimum votes:* `100votes` (optional)
        - *minimum stars:* `100stars` (optional)

        *Tip:* if you want to encourage people to join you, form the query with some popular languages like `javascript python elixir` to increase interest in solving the puzzle. If you want to contribute to the development of this app, join the #engine-room channel.
    """)
