module.exports = (/* TODO options = {} */) => {
    return (req, res, next) => {
        // Just for debugging at the moment
        // console.info("hello world");
        next();
    };
};
