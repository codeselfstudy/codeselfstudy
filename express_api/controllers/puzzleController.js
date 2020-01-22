function index(req, res, next) {
    res.json({ title: "puzzles" });
}

function detail(req, res, next) {
    res.json({ title: `detail for puzzle ${req.params.id}` });
}

module.exports = {
    index,
    detail,
};
