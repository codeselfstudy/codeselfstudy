const express = require("express");
const ctrl = require("../controllers/puzzleController");
const router = express.Router();

router.get("/", ctrl.index);
router.get("/:id", ctrl.detail);

module.exports = router;
