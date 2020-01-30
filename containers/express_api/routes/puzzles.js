const express = require("express");
const ctrl = require("../controllers/puzzleController");
const router = express.Router();

router.get("/:slug", ctrl.detail);
router.get("/", ctrl.index);

module.exports = router;
