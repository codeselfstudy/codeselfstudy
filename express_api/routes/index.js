const express = require("express");
const router = express.Router();
const ctrl = require("../controllers/pageController");

/* GET home page. */
router.get("/", ctrl.index);

module.exports = router;
