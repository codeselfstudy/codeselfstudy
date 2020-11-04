"""
This module will post to Discourse's API.

Documentation for the API:
https://docs.discourse.org/

Use an API key generated specifically for user "content_bot".
"""
import os
import json
import requests
# from urllib.parse import urlencode, quote_plus
# import logging

# logging.basicConfig(filename='app.log', level=logging.INFO, filemode='a', format='%(name)s - %(levelname)s - %(message)s')

DISCOURSE_API_USER = os.getenv("DISCOURSE_API_USER")
DISCOURSE_API_KEY = os.getenv("DISCOURSE_API_KEY")
DISCOURSE_ROOT_URL = os.getenv("DISCOURSE_ROOT_URL")

headers = {
    "Api-Username": DISCOURSE_API_USER,
    "Api-Key": DISCOURSE_API_KEY,
}


def create_forum_post(payload):
    """Create a forum post.

    Sample payload:
    ```
    {
        # these three fields are required
        "title": "string",
        "raw": "string",  # markdown
        "category": 0,    # the puzzle category is stored in an environment var

        # this is only used if posting to an existing topic
        "topic_id": 0,

        # these fields are only for creating private messages
        "target_recipients": "discourse1,discourse2",
        "archetype": "private_message",

        # this is only used if you want the change the `created_at` time
        "created_at": "2017-01-31"
    }
    ```

   The muted bot test category is 44. The puzzles category is 6.

    It will return a response with information about the new topic. You
    can get the `topic_id` and `topic_slug` fields and link to
    f"{DISCOURSE_ROOT_URL}/t/{topic_slug}/{topic_id}"
    """
    endpoint = f"{DISCOURSE_ROOT_URL}/posts.json"

    discourse_response = requests.post(endpoint, headers=headers, data=payload)
    return discourse_response.json()


