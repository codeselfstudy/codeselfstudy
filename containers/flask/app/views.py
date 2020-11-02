import json
from flask import render_template, jsonify, abort, request
from app import app
from . import slack
from . import puzzles as p


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
    # print("data", data)
    if slack.verify_signature(slack_signature, slack_ts, data):
        print("signature valid")

        payload = slack.extract_payload(data)
        if payload:
            query = json.dumps(slack.raw_text_to_query(payload["text"]))
            puzzle = p.query_puzzles(query)
            print("payload", payload)
            print("query", query)
            return jsonify({
                "response_type": "in_channel",
                "blocks": [
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"*{payload['user_name']}* requested a coding puzzle."
                        }
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            # TODO: fix this - move to a function or something
                            "text": f"*{puzzle['name']}*\n- kyu: {puzzle['kyu']}\n- languages: {','.join(puzzle['languages'])}\n- category: {puzzle['category']}\n- url: {puzzle['url']}"
                        }
                    },
                    {
                        "type": "section",
                        "text": {
                            "type": "mrkdwn",
                            "text": f"debug: ```{query}```"
                        }
                    }
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
