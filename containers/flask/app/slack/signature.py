import re
import hmac
import time
import hashlib
from os import environ
from app.helpers.utilities import safe_list_get
from urllib.parse import parse_qs


def verify_signature(slack_signature, ts, request_body):
    """Verifies that a request came from Slack.

    See the documentation here:
    https://api.slack.com/authentication/verifying-requests-from-slack

    also this tutorial:
    https://the-digital-owl.co.uk/2020/05/22/verifying-slack-requests-in-python/
    """
    if (int(time.time()) - int(ts)) > 60:
        print("slack timestamp is too old", ts)
        return False

    SLACK_SIGNING_SECRET = environ.get("SLACK_SIGNING_SECRET")
    secret = bytes(SLACK_SIGNING_SECRET, "utf-8")

    # # Create a basestring by concatenating the version, the request
    #   timestamp, and the request body
    basestring = f"v0:{ts}:{request_body}".encode("utf-8")
    # Hash the basestring using your signing secret, take the hex
    # digest, and prefix with the version number
    my_signature = (
        "v0=" + hmac.new(secret, basestring, hashlib.sha256).hexdigest()
    )
    # # Compare the resulting signature with the signature on the request to verify the request
    if hmac.compare_digest(my_signature, slack_signature):
        return True
    else:
        print("signature invalid")
        return False
