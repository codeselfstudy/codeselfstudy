from os import getenv

DISCOURSE_ROOT_URL = getenv("DISCOURSE_ROOT_URL")


def construct_topic_url(topic_id, topic_slug=None, short=False):
    """Construct a URL to a Discourse topic."""
    if short or not topic_slug:
        return f"{DISCOURSE_ROOT_URL}/t/{topic_id}"
    else:
        return f"{DISCOURSE_ROOT_URL}/t/{topic_slug}/{topic_id}"


# TODO: double-check that this is the right way to do it before uncommenting.
# def construct_post_url(post_id, slug=None):
#     """Construct a URL to a Discourse post."""
#     if slug:
#         return f"{DISCOURSE_ROOT_URL}/p/{slug}/{post_id}"
#     else:
#         return f"{DISCOURSE_ROOT_URL}/p/{post_id}"
