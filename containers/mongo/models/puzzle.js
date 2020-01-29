// quick way to load initial data so we can start using it.
// TODO: this shouldn't duplicate the code from Express. Make npm
// package and install it in both places?
const mongoose = require("mongoose");
const URLSlugs = require("mongoose-url-slugs");
const { clean } = require("../sanitize");
const { db } = require("../db");
require("mongoose-type-url");

const puzzleSchema = new mongoose.Schema({
    title: {
        type: String,
        required: true,
        minlength: 1,
        maxlength: 300,
    },
    source_id: {
        type: String,
        index: true,
        required: true,
    },
    body: {
        type: String,
    },
    unsafe_html: { type: String, required: true, minlength: 5 },
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
        required: true,
        enum: ["hard", "medium", "easy", "unknown"],
    },
});

puzzleSchema.plugin(
    URLSlugs("source title", {
        // This overrides the default slug generation to add a slash
        // after the prefix.
        generator(text, separator) {
            const segments = text.split(" ");
            const slug =
                segments[0] +
                "/" +
                segments
                    .slice(1)
                    .join(" ")
                    .toLowerCase()
                    .replace(/([^a-z0-9\-\_]+)/g, separator)
                    .replace(new RegExp(separator + "{2,}", "g"), separator);
            if (slug.substr(-1) === separator) {
                slug = slug.substr(0, slug.length - 1);
            }
            return slug;
        },
    })
);

// TODO: this might be crashing(?)
puzzleSchema.pre("save", next => {
    this.body = clean(this.unsafe_html);
    next();
});

const Puzzle = mongoose.model("Puzzle", puzzleSchema);

module.exports = { Puzzle };
