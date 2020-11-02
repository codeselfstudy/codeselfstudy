"""Helper functions"""


def safe_list_get(lst, idx, default):
    """A helper function to safely get the first element of a list."""
    try:
        return lst[idx]
    except IndexError:
        return default
