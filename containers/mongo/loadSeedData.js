const fs = require("fs");
const path = require("path");

const DATA_DIR = path.join(__dirname, "project_euler");
const fileList = fs.readdirSync(DATA_DIR).filter(f => /^\d+\.json$/.test(f));

fileList.forEach(f => {
    // save into mongo
    console.log(`processing: ${f}`);
});
