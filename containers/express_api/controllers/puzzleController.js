const { Puzzle } = require("../models/puzzle");

function index(req, res, next) {
    Puzzle.find({})
        .select(["title", "slug", "unsafe_html", "url", "difficulty"])
        .limit(20)
        .exec((err, puzzles) => {
            res.json({ puzzles });
        });
}

function detail(req, res, next) {
    const slug = req.params.slug;
    console.log("=======");
    console.log("slug", slug);
    console.log("url", req.originalUrl);
    console.log("=======");
    Puzzle.findOne({ slug }, (err, puzzle) => {
        if (err) {
            console.error("err", err);
            res.status(404).json({ error: err });
        }
        res.json(puzzle);
    });
    // res.json({ title: `detail for puzzle ${req.params.id} [NOT IMPLEMENTED]` });
}

module.exports = {
    index,
    detail,
};
