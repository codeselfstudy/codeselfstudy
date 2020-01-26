function index(req, res, next) {
    res.json({ title: "hello world" });
}

module.exports = {
    index,
};
