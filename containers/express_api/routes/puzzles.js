const express = require("express");
const ctrl = require("../controllers/puzzleController");
const router = express.Router();

router.get("/:source/:id", ctrl.detail);
router.get("/:source/search", ctrl.search);
router.get("/", ctrl.index);

module.exports = router;
