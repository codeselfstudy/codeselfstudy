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


@app.route("/<source>/<puzzle_id>", defaults={"puzzle_id": None})
def detail(source, puzzle_id):
    """
    This example query gets a random puzzle out of the 700 puzzles that
    have a voteScore of over 400 and that are available in both JS and
    Python.
    ```
    db.puzzles.aggregate([
        {
            $match: {
                source: "codewars",
                voteScore: { $gt: 400 },
                languages: { $all: ["python", "javascript"] },
            },
        },
        { $sample: { size: 1 } },
    ]);
    ```
    """
    if puzzle_id:
        puzzle = mongo.db.puzzles.find_one({"source": source, "id": str(puzzle_id)})
    else:
        puzzle = mongo.db.puzzles.find_one({"source": source})
    if puzzle:
        print("puzzle", puzzle)

# {'_id': ObjectId('5f9f1ff4144695653459dbaf'), 'source': 'codewars', 'id': '50654ddff44f800200000004', 'name': 'Multiply', 'slug': 'multiply', 'category': 'bug_fixes', 'publishedAt': '2013-05-18T18:40:17.975Z', 'approvedAt': '2013-12-03T15:41:04.738Z', 'languages': ['javascript', 'coffeescript', 'ruby', 'python', 'haskell', 'clojure', 'java', 'csharp', 'elixir', 'cpp', 'typescript', 'php', 'crystal', 'dart', 'rust', 'fsharp', 'swift', 'go', 'shell', 'c', 'lua', 'sql', 'bf', 'r', 'nim', 'erlang', 'objc', 'scala', 'kotlin', 'solidity', 'groovy', 'fortran', 'nasm', 'julia', 'powershell', 'purescript', 'elm', 'ocaml', 'reason', 'idris', 'racket', 'agda', 'coq', 'vb', 'forth', 'factor', 'prolog', 'cfml', 'lean', 'cobol', 'haxe', 'commonlisp', 'raku', 'perl'], 'url': 'https://www.codewars.com/kata/50654ddff44f800200000004', 'rank': {'id': -8, 'name': '8 kyu', 'color': 'white'}, 'createdAt': '2012-09-28T07:12:31.171Z', 'approvedBy': {'username': 'alchemy', 'url': 'https://www.codewars.com/users/alchemy'}, 'description': 'This code does not execute properly. Try to figure out why.', 'totalAttempts': 3962887, 'totalCompleted': 3130189, 'totalStars': 1062, 'voteScore': 7822, 'tags': ['Bugs'], 'contributorsWanted': True, 'unresolved': {'issues': 4, 'suggestions': 1}}
        return jsonify({
            "id": puzzle["id"],
            "url": puzzle["url"],
            "name": puzzle["name"],
            "languages": puzzle["languages"],
            "kyu": abs(puzzle["rank"]["id"]),
            "votes": puzzle["voteScore"],
            "stars": puzzle["totalStars"],
            "category": puzzle["category"],
            "tags": puzzle["tags"],
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
