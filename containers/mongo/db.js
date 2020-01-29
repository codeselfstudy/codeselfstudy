const DB_NAME = "puzzles";
const mongoose = require("mongoose");

mongoose.connect(`mongodb://mongo/${DB_NAME}`, {
    useUnifiedTopology: true,
    useNewUrlParser: true,
});

const db = mongoose.connection;

db.on("error", console.error.bind(console, "connection error:"));
db.once("open", () => console.log(`connected to mongodb database: ${DB_NAME}`));

module.exports = {
    db,
};
