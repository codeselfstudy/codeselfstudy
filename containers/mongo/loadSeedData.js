const fs = require("fs");
const path = require("path");
const { db } = require("./db");
const { Puzzle } = require("./models/puzzle");

const DATA_DIR = path.join(__dirname, "seed_data", "project_euler");
const fileList = fs.readdirSync(DATA_DIR).filter(f => /^\d+\.json$/.test(f));

/**
 * Create a path to the downloaded Project Euler problems.
 */
function makePathTo(fname) {
    return path.join(__dirname, "seed_data", "project_euler", fname);
}

/**
 * Process all of the JSON files in the Project Euler directory.
 */
fileList.forEach(f => {
    // save into mongo
    const { id, title, body, url } = JSON.parse(
        fs.readFileSync(makePathTo(f), "utf-8")
    );
    const puzzle = new Puzzle({
        title,
        unsafe_html: body,
        url,
        difficulty: "unknown",
        source_id: id,
        source: "projecteuler",
    });
    console.log(`processing: ${puzzle.title}`);
    puzzle.save();
});

db.close();
