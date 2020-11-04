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
DISCOURSE_PUZZLES_CATEGORY = os.getenv("DISCOURSE_PUZZLES_CATEGORY")

headers = {
    "Api-Username": DISCOURSE_API_USER,
    "Api-Key": DISCOURSE_API_KEY,
}


def create_forum_post(payload):
    """Create a forum post.

    Sample payload:
    ```
    {
        "title": "string",
        "raw": "string",
        "category": 0,
    }
    ```

   The muted bot test category is 44. The puzzles category is 6.

    It will return a response with information about the new topic. You
    can get the `topic_id` and `topic_slug` fields and link to
    f"{DISCOURSE_ROOT_URL}/t/{topic_slug}/{topic_id}"
    """
    # params = {}
    # param_string = urlencode(params, quote_via=quote_plus)
    # endpoint = "{}/posts.json?{}".format(BASE_URL, param_string)
    endpoint = f"{DISCOURSE_ROOT_URL}/posts.json"
    # logging.info(f"making request to {endpoint}")

    res = requests.post(endpoint, headers=headers, data=payload)
    t = res.json()
    # print(json.dumps(t, indent=4))
    # topic_url = f"{DISCOURSE_ROOT_URL}/t/{t.topic_slug}/{t.topic_id}"
    return t


