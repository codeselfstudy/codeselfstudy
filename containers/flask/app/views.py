import json
from flask import render_template, jsonify, abort, request
from app import app
from app.slack import puzzle_command, signature
import app.puzzles as p
from app.helpers import format_codewars_puzzle_message


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
                    "text": "Your query wasn't formatted correctly or there was another error. Try typing something like this:\n```/puzzle codewars js python elixir 5kyu```\n\nParameters can look like this:\n- languages: `python javascript fortran` (optional)\n- difficulty: `5kyu` (optional)\n- minimum votes: `100votes` (optional)\n- minimum stars: `100stars` (optional)"
                })
            else:
                print("puzzle data", puzzle)
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
                                "text": format_codewars_puzzle_message(puzzle)
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
