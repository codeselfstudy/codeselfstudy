const express = require("express");
const ctrl = require("../controllers/authController");
const router = express.Router();

// Request a Discourse SSO URL
router.get("/request", ctrl.request);

// Validate the sign-on
router.get("/validate", ctrl.validate);

// Return a JWT if one exists in the session
router.post("/jwt", ctrl.jwt);

// Return the logged-in status of the current session
router.post("/check", ctrl.check);

// Log out of Discourse
router.post("/logout", ctrl.logout);

// Catch remaining GET requests with a message for developers. They
// might be trying to use GET requests instead of POST requests.
router.get("*", ctrl.lost);

module.exports = router;
