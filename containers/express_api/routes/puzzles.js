const express = require("express");
const ctrl = require("../controllers/puzzleController");
const router = express.Router();

router.get("/:slug", ctrl.detail);
router.get("/", ctrl.index);

router.post("/:slug/save", ctrl.save);

module.exports = router;
