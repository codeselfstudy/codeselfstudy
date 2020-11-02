from pymongo import MongoClient
from flask import render_template, jsonify, abort
from flask_pymongo import PyMongo
from app import app

app.config["MONGO_URI"] = "mongodb://mongo/codeselfstudy"
mongo = PyMongo(app)


def safe_list_get(lst, idx, default):
    """A helper function to safely get the first element of a list."""
    try:
        return lst[idx]
    except IndexError:
        return default


@app.route("/")
def home():
    return render_template("home.html")


@app.route("/puzzles/<source>", defaults={"puzzle_id": None})
@app.route("/puzzles/<source>/<puzzle_id>")
def detail(source, puzzle_id):
    if puzzle_id:
        puzzle = mongo.db.puzzles.find_one({"source": source, "id": str(puzzle_id)})
    else:
        # This query gets a random puzzle out of the 700 puzzles that
        # have a voteScore of over 400 and that are available in both JS
        # and Python. Adjust as desired.
        q = list(mongo.db.puzzles.aggregate([
            {
                "$match": {
                    "source": "codewars",
                    "voteScore": {"$gt": 400},
                    "languages": {"$all": ["python", "javascript"]},
                    },
                },
            {"$sample": {"size": 1}},
            ]))

        print("q", q)
        puzzle = safe_list_get(q, 0, None)

    # return the puzzle if it was found, otherwise return a 404
    if puzzle:
        return jsonify({
            "id": puzzle.get("id", None),
            "url": puzzle.get("url", None),
            "name": puzzle.get("name", None),
            "languages": puzzle.get("languages", None),
            "kyu": abs(puzzle.get("rank", None).get("id", None)),
            "votes": puzzle.get("voteScore", None),
            "stars": puzzle.get("totalStars", None),
            "category": puzzle.get("category", None),
            "tags": puzzle.get("tags", None),
        })
    else:
        abort(404)


@app.route("/puzzles/search/codewars")
def search_codewars():
    # languages = request.args.get("languages").split(",")
    # db_query = collection.find({
    #     "languages"
    # })
    return jsonify({"msg": "not implemented"})


@app.route("/puzzles/search/projecteuler")
def search_projecteuler():
    return jsonify({"msg": "not implemented"})


@app.route("/puzzles/search/leetcode")
def search_leetcode():
    return jsonify({"msg": "not implemented"})
