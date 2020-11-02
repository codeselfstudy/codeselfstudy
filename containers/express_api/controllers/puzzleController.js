const { db } = require("../db/mongo");

function index(req, res) {
    res.json({ msg: "Coming soon" });
}

function search(req, res) {
    // const params = req.params;
    // const { kyu, minvotes, minstars, languages, title } = req.query;
    // const result = db.puzzles.findOne();
    res.json({ msg: "search coming soon" });
}

function detail(req, res) {
    console.log("db", db);
    const params = req.params;
    const result = db.puzzles.findOne();
    res.json({ msg: "search coming soon", result: result });
    res.json({ msg: "Coming soon" });
}

module.exports = {
    index,
    detail,
    search,
};
