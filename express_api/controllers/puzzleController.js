function index(req, res, next) {
    res.json({ title: "puzzles [NOT IMPLEMENTED]" });
}

function detail(req, res, next) {
    res.json({ title: `detail for puzzle ${req.params.id} [NOT IMPLEMENTED]` });
}

module.exports = {
    index,
    detail,
};
