function request(req, res, next) {
    res.json({ title: `request auth` });
}

function validate(req, res, next) {
    res.json({ title: `validate auth` });
}

module.exports = {
    request,
    validate,
};
