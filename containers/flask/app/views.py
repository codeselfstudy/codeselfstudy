from pymongo import MongoClient
from flask import render_template, jsonify, abort
from app import app

client = MongoClient("mongo", 27017)
db = client["codeselfstudy"]
collection = db["puzzles"]
# client = pymongo.MongoClient("mongodb://localhost:27017/")


@app.route("/")
def home():
    return render_template("home.html")


@app.route("/<source>/<puzzle_id>", defaults={"puzzle_id": None})
def detail(source, puzzle_id):
    if puzzle_id:
        puzzle = collection.find_one({"source": source, "id": puzzle_id})
    else:
        puzzle = collection.find_one({"source": source})
    if puzzle:
        return jsonify(puzzle)
    else:
        abort(404)


@app.route("/search/codewars")
def search_codewars():
    # languages = request.args.get("languages").split(",")
    # db_query = collection.find({
    #     "languages"
    # })
    pass


@app.route("/search/projecteuler")
def search_projecteuler():
    pass


@app.route("/search/leetcode")
def search_leetcode():
    pass
