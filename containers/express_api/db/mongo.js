const DB_NAME = "puzzles";
const { MongoClient } = require("mongodb");

const uri = `mongodb://mongo/`;
let db;

MongoClient.connect(uri, (err, database) => {
    if (err) {
        database.close();
        throw err;
    } else {
        db = database.db(DB_NAME);
    }
});

// `db.collection("challenges").findOnde({}, (err, result) => {
//  if (err) { throw err; }
//  do something with result here
// })`
module.exports = {
    db,
};
