const discourse = require("../auth/discourse");

/**
 * Generate an SSO URL and redirect the user to Discourse.
 */
const request = (req, res, next) => {
    // If there is a destination URL in the URL parameters, save it.
    const { destination } = req.query;
    req.session.destination = destination;

    const { nonce, destinationUrl } = discourse.generateAuthUrl({
        SSO_PROVIDER_SECRET: process.env.SSO_PROVIDER_SECRET,
        RETURN_URL: process.env.RETURN_URL,
        DISCOURSE_ROOT_URL: process.env.DISCOURSE_ROOT_URL,
    });

    req.session.nonce = nonce;
    req.session.save();

    res.json({ url: destinationUrl });
};

/**
 * Handle the information that Discourse passes back in a querystring.
 */
const validate = (req, res, next) => {
    const { sso, sig } = req.query;

    if (req.session.nonce) {
        const result = discourse.validateAuth(sso, sig, req.session.nonce, {
            SSO_PROVIDER_SECRET: process.env.SSO_PROVIDER_SECRET,
        });

        const destination = req.session.destination || process.env.BASE_URL;

        const discourseUid = result.payload.external_id;
        const additionalClaims = {
            discourseUid: discourseUid,
            username: result.payload.username, // the canonical username
            name: result.payload.name || "", // an optional full name
            isAdmin: result.payload.admin || false,
            avatarUrl: result.payload.avatar_url || "",
            emailAddress: result.payload.email.toLowerCase(),
        };

        req.session.claims = additionalClaims;

        // console.log('#######');
        // doesUserExist('3').then(x => console.log('x', x));

        // TODO: create the session here.
        // req.session.whatever = whatever;
        // res.redirect("somewhere?msg=success");
    } else {
        res.json({ msg: "there was a problem" });
    }
};

/**
 * Returns JSON Web Token(s)
 *
 * TODO: are we going to issue JWT? This is from other code.
 */
const jwt = (req, res, next) => {
    const token = req.session.jwt || null;
    res.json({
        msg: "[[NOT IMPLEMENTED]]",
        token,
    });
};

/**
 * Lets requester know whether the user is logged in.
 *
 * Sends a Discourse user ID if authenticated, otherwise a 401 header.
 */
const check = (req, res, next) => {
    const sess = req.session;
    if (sess.claims && sess.claims.discourseUid) {
        res.json({ status: true, uid: sess.claims.discourseUid });
    } else {
        res.send(401, { status: false, uid: null });
    }
};

/**
 * Log out of Discourse
 */
const logout = (req, res, next) => {
    const discourseUid =
        (req.session &&
            req.session.claims &&
            req.session.claims.discourseUid) ||
        null;

    req.session.destroy();
    res.json({ action: "logout" });

    // It's okay to run a function after the response is sent.
    if (discourseUid) {
        console.log(`logging out user with Discourse ID ${discourseUid}`);
        discourse.logout(discourseUid);
        console.log(
            `post discourse logout -- session is: ${JSON.stringify(
                req.session
            )}`
        );
    }
};

module.exports = {
    request,
    validate,
    jwt,
    check,
    logout,
};
