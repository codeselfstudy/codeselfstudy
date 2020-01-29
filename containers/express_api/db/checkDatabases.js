const { Puzzle } = require("../models/puzzle");

/**
 * Check to see whether data has already been seeded in the database.
 *
 * For the moment, only some things will be seeded during build, like
 * Project Euler problems. Once users submit data, the database will
 * probably not get seeded like this.
 */
function isDatabaseSeeded({ projectEuler }) {
    if (projectEuler === true) {
        Puzzle.findOne({ source_id: "projecteuler" }, (err, puzzle) => {
            if (err) console.log("err", err);

            console.log(`found puzzle`, puzzle);
        });
    }

    return {
        projectEuler,
    };
}

module.exports = {
    isDatabaseSeeded,
};
