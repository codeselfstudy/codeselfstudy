const mongoose = require("mongoose");
const { db } = require("../db/mongo");
require("mongoose-type-url");

const puzzleSchema = new mongoose.Schema({
    title: {
        type: String,
        required: true,
        minlength: 10,
        maxlength: 300,
    },
    source_id: {
        type: String,
    },
    body: { type: String, required: true, minlength: 25 },
    unsafe_html: { type: String, required: true },
    source: {
        type: String,
        required: true,
        enum: ["codewars", "projecteuler", "leetcode", "self"],
    },
    // The URL that points to the source of a puzzle
    url: {
        type: mongoose.SchemaTypes.Url,
    },
    difficulty: {
        type: String,
        enum: ["hard", "medium", "easy", "unknown"],
    },
});

const Puzzle = mongoose.model("Puzzle", puzzleSchema);

module.exports = { Puzzle };
