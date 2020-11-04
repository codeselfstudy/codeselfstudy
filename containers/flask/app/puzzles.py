"""This module gets puzzles out of the database."""
from flask_pymongo import PyMongo
from app import app
from app.helpers.formatters import safe_list_get

app.config["MONGO_URI"] = "mongodb://mongo/codeselfstudy"
mongo = PyMongo(app)


def get_puzzle(source, puzzle_id=None):
    if source == "codewars":
        return _get_codewars_puzzle(puzzle_id)
    elif source == "projecteuler":
        return _get_projecteuler_puzzle(puzzle_id)
    elif source == "leetcode":
        return _get_leetcode_puzzle(puzzle_id)
    else:
        return None


def query_puzzles(query):
    print("query", query)
    if query and query.get("source", None) == "codewars":
        return _query_codewars_puzzle(query)
    # elif source == "projecteuler":
    #     return _query_projecteuler_puzzle(query)
    # elif source == "leetcode":
    #     return _query_leetcode_puzzle(query)
    else:
        return None


def _query_codewars_puzzle(query):
    q = {
        "$match": {
            "source": "codewars",
        },
    }
    # {"source": "codewars", "languages": ["python"], "kyu": 4, "stars": 100}
    if query.get("votes", None):
        q["$match"]["voteScore"] = {"$gt": query["votes"]}
    if query.get("kyu", None):
        q["$match"]["rank.id"] = -query["kyu"]
    if query.get("stars", None):
        q["$match"]["totalStars"] = {"$gt": query["votes"]}
    if len(query["languages"]) > 0:
        q["$match"]["languages"] = {"$all": query["languages"]}

    result = list(mongo.db.puzzles.aggregate([
        q,
        {"$sample": {"size": 1}},
    ]))

    print("q", q)
    print("result", result)
    puzzle = safe_list_get(result, 0, None)

    # TODO: refactor this, because it's a repeat of some code below
    if puzzle:
        return {
            "id": puzzle.get("id", None),
            "url": puzzle.get("url", None),
            "name": puzzle.get("name", None),
            "languages": puzzle.get("languages", None),
            "kyu": abs(puzzle.get("rank", None).get("id", None)),
            "votes": puzzle.get("voteScore", None),
            "stars": puzzle.get("totalStars", None),
            "category": puzzle.get("category", None),
            "tags": puzzle.get("tags", None),
        }
    else:
        return None


def _search_codewars_puzzles(query_dict):
    pass


def _get_projecteuler_puzzle(puzzle_id):
    # TODO: implement this
    return None


def _get_leetcode_puzzle(puzzle_id):
    # TODO: implement this
    return None


def _get_codewars_puzzle(puzzle_id):
    if puzzle_id:
        puzzle = mongo.db.puzzles.find_one({"source": "codewars", "id": str(puzzle_id)})
    else:
        # This query gets a random puzzle out of the 700 codewars
        # puzzles that have a voteScore of over 400 and that are
        # available in both JS and Python. Adjust as desired.
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
    if puzzle:
        return {
            "id": puzzle.get("id", None),
            "url": puzzle.get("url", None),
            "name": puzzle.get("name", None),
            "languages": puzzle.get("languages", None),
            "kyu": abs(puzzle.get("rank", None).get("id", None)),
            "votes": puzzle.get("voteScore", None),
            "stars": puzzle.get("totalStars", None),
            "category": puzzle.get("category", None),
            "tags": puzzle.get("tags", None),
        }
    else:
        return None
