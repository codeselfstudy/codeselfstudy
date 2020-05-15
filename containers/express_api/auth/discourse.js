/**
 * This module contains functionality for using Discourse as an SSO
 * provider.
 */
const querystring = require("querystring");
const crypto = require("crypto");
const axios = require("axios");
const DISCOURSE_API_KEY = process.env.DISCOURSE_API_KEY;
const DISCOURSE_API_USER = process.env.DISCOURSE_API_USER;

/**
 * Generate an authorization URL.
 *
 * The options object should contain:
 *
 *  - the SSO_PROVIDER_SECRET from the .env file
 *  - the RETURN_URL (where to go after authenticating)
 *  - the DISCOURSE_ROOT_URL
 */
const generateAuthUrl = options => {
    // Generate a random nonce and save it temporarily.
    // The nonce needs to be persisted for a short time, so we should
    // save the creation time.
    const nonce = {
        value: crypto.randomBytes(16).toString("hex"),
        createdAt: new Date(),
    };

    // create payload with nonce and return url
    //      nonce=nonce&return_sso_url=return_url
    const qs = `nonce=${nonce.value}&return_sso_url=${options.RETURN_URL}`;

    // base64 encode the PAYLOAD
    const qsB64 = new Buffer.from(qs).toString("base64");

    // Generate a HMAC-SHA256 from the base 64 payload using the
    // SSO_PROVIDER_SECRET as the key.
    const hmac = crypto.createHmac("sha256", options.SSO_PROVIDER_SECRET);
    hmac.update(qsB64);

    // create a lowercase hex string from that
    const sig = hmac
        .digest("hex")
        .split("")
        .map(char => char.toLowerCase())
        .join("");

    // URL encode the base 64 payload
    const qsURIencoded = encodeURIComponent(qsB64);

    // redirect the user to DISCOURSE_ROOT_URL/session/sso_provider? \
    //     sso=URL_ENCODED_PAYLOAD&sig=HEX_SIGNATURE
    const query = `sso=${qsURIencoded}&sig=${sig}`;

    const destinationUrl = `${options.DISCOURSE_ROOT_URL}/session/sso_provider?${query}`;

    return {
        nonce, // an object to be persisted for a few minutes
        destinationUrl, // a string
    };
};

/**
 * Extract the auth querystring from the URL and validate it.
 */
const validateAuth = (sso, sig, nonceObj, options) => {
    const hmac = crypto.createHmac("sha256", options.SSO_PROVIDER_SECRET);
    const decoded = decodeURIComponent(sso);
    hmac.update(decoded);
    const hash = hmac.digest("hex");

    let result;

    if (sig === hash && _nonceTimeIsValid(nonceObj)) {
        const userObj = new Buffer.from(sso, "base64").toString("utf8");

        // console.log('userObj', userObj);

        const payload = querystring.parse(userObj);

        // Make sure the payload.nonce from Discourse === nonceObj.value from Redis
        if (payload.nonce === nonceObj.value) {
            // user can now be logged in
            // console.log('payload', payload);

            // send the payload back to where it can be saved in the session
            return { isValid: true, payload };
        } else {
            // nonce isn't valid
            return {
                isValid: false,
                msg:
                    "Your authorization request timed out. Please try again or contact us.",
            };
        }
    } else {
        // signature and/or nonce not valid
        return {
            isValid: false,
            msg: "did not work",
        };
    }
};

/**
 * Log the given user out of Discourse.
 *
 * TODO: should this be async? The response doesn't need to be held up.
 */
function logout(discourseUid) {
    if (!discourseUid) {
        return;
    }

    const logoutUrl = `https://forum.codeselfstudy.com/admin/users/${discourseUid}/log_out`;
    const headers = {
        "Api-Key": DISCOURSE_API_KEY,
        "Api-Username": DISCOURSE_API_USER,
    };
    const options = {
        url: logoutUrl,
        method: "POST",
        headers,
    };
    // TODO: send API request to Discourse.
    setTimeout(() => {
        axios(options).then(res => console.log("axios response", res));
    }, 2000);
    return;
}

/**
 * Ensure that the nonce was created within the last 90 seconds (or
 * whatever).
 */
function _nonceTimeIsValid(nonceObj, ms = 90000) {
    return new Date() - new Date(nonceObj.createdAt) < ms;
}

module.exports = {
    generateAuthUrl,
    validateAuth,
    logout,
};
