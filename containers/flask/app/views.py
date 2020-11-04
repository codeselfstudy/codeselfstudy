import json
from flask import render_template, jsonify, abort, request
from app import app
from app.slack import puzzle_command, signature
import app.puzzles as p
from app.helpers.formatters import format_codewars_puzzle_message, format_slack_error_message, format_codewars_puzzle_for_discourse
from app.discourse.api import create_forum_post
from app.discourse.helpers import construct_topic_url


@app.route("/")
def home():
    return render_template("home.html")


@app.route("/puzzles", defaults={"source": None, "puzzle_id": None})
@app.route("/puzzles/<source>", defaults={"puzzle_id": None})
@app.route("/puzzles/<source>/<puzzle_id>")
def detail(source, puzzle_id):
    """This sends back a puzzle by source and ID, otherwise a random puzzle."""
    puzzle = p.get_puzzle(source, puzzle_id)
    if puzzle:
        return jsonify(puzzle)
    else:
        abort(404)


@app.route("/puzzles/slack", methods=["POST"])
def slack_slash_command():
    slack_signature = request.headers.get("X-Slack-Signature")
    slack_ts = request.headers.get("X-Slack-Request-Timestamp")
    data = request.get_data().decode()
    print("data in controller", data)
    if signature.verify_signature(slack_signature, slack_ts, data):
        print("signature valid")

        payload = puzzle_command.extract_payload(data)
        if payload:
            query = puzzle_command.raw_text_to_query(payload["text"])
            print("query in the controller", query)
            puzzle = p.query_puzzles(query)
            print("payload", payload)
            print("query", query)
            if not puzzle:
                return jsonify({
                    "text": format_slack_error_message()
                })
            else:
                print("puzzle data", puzzle)
                forum_post_data = format_codewars_puzzle_for_discourse(puzzle)
                print("forum post data", forum_post_data)
                discourse_response = create_forum_post(forum_post_data)
                print("discourse response", discourse_response)

                link_to_forum_topic = construct_topic_url(discourse_response.get("topic_id", None), topic_slug=discourse_response.get("topic_slug", None), short=True)

                body = format_codewars_puzzle_message(puzzle)
                body += f"\n\nForum discussion:\n{link_to_forum_topic}"

                return jsonify({
                    "response_type": "in_channel",
                    "blocks": [
                        # {
                        #     "type": "section",
                        #     "text": {
                        #         "type": "mrkdwn",
                        #         "text": f"*{payload['user_name']}* requested a coding puzzle."
                        #     }
                        # },
                        {
                            "type": "section",
                            "text": {
                                "type": "mrkdwn",
                                "text": body
                            }
                        },
                    ]
                })
        else:
            print("####### there wasn't a payload")
            abort(404)
    else:
        print("####### signature invalid")
        abort(404)


# @app.route("/puzzles/search/codewars")
# def search_codewars():
#     return jsonify({"msg": "not implemented"})


# @app.route("/puzzles/search/projecteuler")
# def search_projecteuler():
#     return jsonify({"msg": "not implemented"})


# @app.route("/puzzles/search/leetcode")
# def search_leetcode():
#     return jsonify({"msg": "not implemented"})
