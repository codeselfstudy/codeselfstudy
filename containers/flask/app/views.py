from pymongo import MongoClient
from flask import render_template, jsonify, abort
from flask_pymongo import PyMongo
from app import app

app.config["MONGO_URI"] = "mongodb://mongo/codeselfstudy"
mongo = PyMongo(app)

# client = MongoClient("mongo", 27017)
# db = client["codeselfstudy"]
# collection = db["puzzles"]
# client = pymongo.MongoClient("mongodb://localhost:27017/")
# r = collection.find_one()
# print("#######", r)


@app.route("/")
def home():
    return render_template("home.html")


@app.route("/<source>", defaults={"puzzle_id": None})
@app.route("/<source>/<puzzle_id>")
def detail(source, puzzle_id):
    if puzzle_id:
        puzzle = mongo.db.puzzles.find_one({"source": source, "id": str(puzzle_id)})
    else:
        # This query gets a random puzzle out of the 700 puzzles that
        # have a voteScore of over 400 and that are available in both JS
        # and Python. Adjust as desired.
        puzzle = list(mongo.db.puzzles.aggregate([
            {
                "$match": {
                    "source": "codewars",
                    "voteScore": {"$gt": 400},
                    "languages": {"$all": ["python", "javascript"]},
                    },
                },
            {"$sample": {"size": 1}},
            ]))[0]

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


@app.route("/search/codewars")
def search_codewars():
    # languages = request.args.get("languages").split(",")
    # db_query = collection.find({
    #     "languages"
    # })
    return jsonify({"msg": "not implemented"})


@app.route("/search/projecteuler")
def search_projecteuler():
    return jsonify({"msg": "not implemented"})


@app.route("/search/leetcode")
def search_leetcode():
    return jsonify({"msg": "not implemented"})
