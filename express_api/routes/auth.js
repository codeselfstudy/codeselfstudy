const express = require("express");
const ctrl = require("../controllers/authController");
const router = express.Router();

router.get("/request", ctrl.request);
router.get("/validate", ctrl.validate);

module.exports = router;
