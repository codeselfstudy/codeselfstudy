const sanitizeHtml = require("sanitize-html");

// That will allow our default list of allowed tags and attributes through. It's a nice set, but probably not quite what you want. So:

// Allow only a super restricted set of tags and attributes
const clean = dirty =>
    sanitizeHtml(dirty, {
        allowedTags: [
            "p",
            "b",
            "i",
            "em",
            "strong",
            "sup",
            "sub",
            "blockquote",
            "li",
            "ol",
            "ul",
            // "img",
            // "a",
            "div",
            "table",
            "tr",
            "th",
            "td",
            "tbody",
            "thead",
            "span",
            "br",
            "var",
        ],
        allowedAttributes: {
            // the links will be broken until they are rewritten
            // a: ["href"],
            // img: ["src", "alt"],
            span: ["style"],
            div: ["style"],
        },
    });

module.exports = {
    clean,
};
