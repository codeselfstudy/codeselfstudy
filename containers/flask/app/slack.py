"""
This module handles Slack integration.
"""
import re
import hmac
import time
import hashlib
from os import environ
from .helpers import safe_list_get
from urllib.parse import parse_qs

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


# We'll probably want these items:
# user_name=josh
# text=codewars+6kyu+javascript+elixir+python+200votes+10stars <- parse this
# api_app_id=os.environ.get("SLACK_APP_ID")


def extract_payload(payload):
    """Extract the form encoded data from slack into a dict.

    This also checks that it was sent by the correct (our) slack app.
    """
    data = parse_qs(payload)
    if data["api_app_id"][0] == environ.get("SLACK_APP_ID"):
        result = {
            "user_id": safe_list_get(data.get("user_id", None), 0, None),
            "user_name": safe_list_get(data.get("user_name", None), 0, None),
            "text": safe_list_get(data.get("text", None), 0, None),
            "response_url": safe_list_get(data.get("response_url", None), 0, None),
            "channel_name": safe_list_get(data.get("channel_name", None), 0, None),
            "channel_id": safe_list_get(data.get("channel_id", None), 0, None),
            "command": safe_list_get(data.get("command", None), 0, None),
        }
        print("result is", result)
        return result
    else:
        print(f"the payload has the wrong app id: {data['api_app_id']}")
        return None


def raw_text_to_query(text):
    """Turns the `text` field of a Slack message into a query."""
    # extract words and remove empty spaces
    words = [w.strip().lower() for w in text.split(" ") if w.strip()]

    # TODO: if the first word is "help" then it should return help information

    # figure out what kind of mongo query should be generated
    sites = [
        "codewars",
        # TODO enable these
        # "projecteuler",
        # "leetcode",
    ]
    site = None
    for s in sites:
        if s in words:
            site = s

    if site == "codewars":
        return _generate_codewars_query(words)
    # elif site == "leetcode":
    #     return _generate_leetcode_query(words)
    # elif site == "projecteuler":
    #     return _generate_projecteuler_query(words)

    return None


def _generate_codewars_query(words):
    """Generates a mongo query based on the text."""
    # these were extracted from the mongo database
    # it's a dict for faster lookup
    valid_languages = {
        "vb": True,
        "forth": True,
        "swift": True,
        "commonlisp": True,
        "cpp": True,
        "powershell": True,
        "prolog": True,
        "cfml": True,
        "lua": True,
        "haxe": True,
        "javascript": True,
        "kotlin": True,
        "php": True,
        "nasm": True,
        "factor": True,
        "agda": True,
        "haskell": True,
        "java": True,
        "erlang": True,
        "clojure": True,
        "elm": True,
        "idris": True,
        "scala": True,
        "solidity": True,
        "r": True,
        "groovy": True,
        "fortran": True,
        "ocaml": True,
        "crystal": True,
        "shell": True,
        "racket": True,
        "nim": True,
        "purescript": True,
        "cobol": True,
        "elixir": True,
        "lean": True,
        "julia": True,
        "rust": True,
        "dart": True,
        "sql": True,
        "bf": True,
        "csharp": True,
        "raku": True,
        "perl": True,
        "typescript": True,
        "go": True,
        "reason": True,
        "objc": True,
        "python": True,
        "coffeescript": True,
        "ruby": True,
        "fsharp": True,
        "coq": True,
        "c": True,
    }
    query = {
        "languages": []
    }
    pattern = re.compile(r"^(\d{1,2})(kyu|votes?|stars?)$")

    for w in words:
        m = re.match(pattern, w)
        if m:
            # example:
            # level: "6", key: "kyu"
            level, key = m.groups()
            query[key] = level
        else:
            if valid_languages.get(w, None):
                query["languages"].append(w)

    print("generated query", query)
    return query


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
