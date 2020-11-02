"""
This module handles Slack integration.
"""
import hmac
import time
import hashlib
from os import environ
# from urllib.parse import parse_qs

# Example payload:
#
# token=gIkuvaNzQIHg97ATvDxqgjtO <- don't use this
# &team_id=T0001
# &team_domain=example
# &enterprise_id=E0001
# &enterprise_name=Globular%20Construct%20Inc
# &channel_id=C2147483705
# &channel_name=test
# &user_id=U2147483697
# &user_name=Steve
# &command=/weather
# &text=94070 <- this will be the payload
# &response_url=https://hooks.slack.com/commands/1234/5678
# &trigger_id=13345224609.738474920.8088930838d88f008e0
# &api_app_id=A123456


def verify_signature(slack_signature, ts, request_body):
    """Verifies that a request came from Slack.

    See the documentation here:
    https://api.slack.com/authentication/verifying-requests-from-slack
    """
    if (int(time.time()) - int(ts)) > 60:
        print("slack timestamp is too old", ts)
        return False

    SLACK_SIGNING_SECRET = environ("SLACK_SIGNING_SECRET")
    # secret = bytes(SLACK_SIGNING_SECRET, "utf-8")

    # # Create a basestring by concatenating the version, the request
    #   timestamp, and the request body
    basestring = f"v0:{ts}:{request_body}".encode("utf-8")
    # Hash the basestring using your signing secret, take the hex
    # digest, and prefix with the version number
    my_signature = (
        "v0=" + hmac.new(SLACK_SIGNING_SECRET, basestring, hashlib.sha256).hexdigest()
    )
    # # Compare the resulting signature with the signature on the request to verify the request
    if hmac.compare_digest(my_signature, slack_signature):
        return True
    else:
        print("signature invalid")
        return False
