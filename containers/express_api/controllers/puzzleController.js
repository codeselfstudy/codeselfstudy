const db = require("../db/mongo");

function index(req, res, _next) {
    // Puzzle.find({})
    //     .select(["title", "slug", "unsafe_html", "url", "difficulty"])
    //     .limit(20)
    //     .exec((err, puzzles) => {
    //         res.json({ puzzles });
    //     });
    res.json({ msg: "Coming soon" });
}

function search(req, res) {
    res.json({ msg: "search coming soon" });
}

function detail(req, res, _next) {
    // const slug = req.params.slug;
    // console.log("=======");
    // console.log("slug", slug);
    // console.log("url", req.originalUrl);
    // console.log("=======");
    // Puzzle.findOne({ slug: slug }, (err, puzzle) => {
    //     if (err) {
    //         console.error("err", err);
    //         res.status(404).json({ error: err });
    //     }
    //     res.json(puzzle);
    // });
    // res.json({ title: `detail for puzzle ${req.params.id} [NOT IMPLEMENTED]` });
    res.json({ msg: "Coming soon" });
}

module.exports = {
    index,
    detail,
};
