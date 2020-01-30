const { Puzzle } = require("../models/puzzle");

/**
 * Check to see whether data has already been seeded in the database.
 *
 * For the moment, only some things will be seeded during build, like
 * Project Euler problems. Once users submit data, the database will
 * probably not get seeded like this.
 *
 * It should return `true` or `false` depending on whether the seeder
 * should be run for that data set.
 */
function isDatabaseSeeded({
    projectEuler /* add more sites here if necessary */,
}) {
    let shouldSeedProjectEuler = false;

    if (projectEuler === true) {
        // If there is at least one Project Euler problem, then don't
        // seed the database.
        Puzzle.findOne({ source: "projecteuler" }, (err, puzzle) => {
            if (err) console.log("err", err);

            if (puzzle) {
                console.log(`project euler database was already seeded`);
            } else {
                shouldSeedProjectEuler = true;
            }
        });
    }

    return {
        shouldSeedProjectEuler, // `true` if the seeder should be run
    };
}

module.exports = {
    isDatabaseSeeded,
};
