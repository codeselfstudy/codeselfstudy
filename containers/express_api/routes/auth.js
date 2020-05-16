const express = require("express");
const ctrl = require("../controllers/authController");
const router = express.Router();

// Return a Discourse SSO URL
router.post('/request', ctrl.authenticate);

// Validate the sign-on
router.get('/validate', ctrl.validate);

// Return a JWT if one exists in the session
router.post('/jwt', ctrl.jwt);

// Return the logged-in status of the current session
router.post('/check', ctrl.check);

// Log out of Discourse
router.post('/logout', ctrl.logout);

module.exports = router;
