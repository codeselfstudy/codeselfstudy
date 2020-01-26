function request(req, res, next) {
    res.json({ title: `request auth [NOT IMPLEMENTED]` });
}

function validate(req, res, next) {
    res.json({ title: `validate auth [NOT IMPLEMENTED]` });
}

module.exports = {
    request,
    validate,
};
