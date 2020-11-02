from flask import render_template
from app import app


@app.route("/")
def home():
    print("hit flask")
    return render_template("home.html")
