// These rules will be written to a robots.txt file when the site is built.
module.exports = {
    host: null,
    sitemap: null,
    policy: [
        {
            userAgent: "ia_archiver",
            disallow: "/",
        },
        {
            userAgent: "archive.org_bot",
            disallow: "/",
        },
        {
            userAgent: "*",
            allow: ["/"],
            disallow: ["/cdn-cgi/", "/api/"],
        },
    ],
};
