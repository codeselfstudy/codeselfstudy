function index(req, res, next) {
    res.json({ title: "home" });
}

module.exports = {
    index
};
