module.exports = {
    siteMetadata: {
        title: `Code Self Study`,
        description: `Programming community in Berkeley, California`,
        author: `@codeselfstudy`,
    },
    plugins: [
        `gatsby-plugin-sass`,
        `gatsby-transformer-toml`,
        `gatsby-plugin-react-helmet`,
        {
            resolve: `gatsby-remark-prismjs`,
            options: {
                // Class prefix for <pre> tags containing syntax highlighting;
                // defaults to 'language-' (e.g. <pre class="language-js">).
                // If your site loads Prism into the browser at runtime,
                // (e.g. for use with libraries like react-live),
                // you may use this to prevent Prism from re-processing syntax.
                // This is an uncommon use-case though;
                // If you're unsure, it's best to use the default value.
                classPrefix: "language-",
                // This is used to allow setting a language for inline code
                // (i.e. single backticks) by creating a separator.
                // This separator is a string and will do no white-space
                // stripping.
                // A suggested value for English speakers is the non-ascii
                // character '›'.
                inlineCodeMarker: null,
                // This lets you set up language aliases.  For example,
                // setting this to '{ sh: "bash" }' will let you use
                // the language "sh" which will highlight using the
                // bash highlighter.
                aliases: {},
                // This toggles the display of line numbers globally alongside the code.
                // To use it, add the following line in gatsby-browser.js
                // right after importing the prism color scheme:
                //  require("prismjs/plugins/line-numbers/prism-line-numbers.css")
                // Defaults to false.
                // If you wish to only show line numbers on certain code blocks,
                // leave false and use the {numberLines: true} syntax below
                showLineNumbers: false,
                // If setting this to true, the parser won't handle and highlight inline
                // code used in markdown i.e. single backtick code like `this`.
                noInlineHighlight: false,
                // This adds a new language definition to Prism or extend an already
                // existing language definition. More details on this option can be
                // found under the header "Add new language definition or extend an
                // existing language" below.
                languageExtensions: [
                    {
                        language: "superscript",
                        extend: "javascript",
                        definition: {
                            superscript_types: /(SuperType)/,
                        },
                        insertBefore: {
                            function: {
                                superscript_keywords: /(superif|superelse)/,
                            },
                        },
                    },
                ],
                // Customize the prompt used in shell output
                // Values below are default
                prompt: {
                    user: "root",
                    host: "localhost",
                    global: false,
                },
                // By default the HTML entities <>&'" are escaped.
                // Add additional HTML escapes by providing a mapping
                // of HTML entities and their escape value IE: { '}': '&#123;' }
                escapeEntities: {},
            },
        },
        {
            resolve: `gatsby-source-filesystem`,
            options: {
                name: `images`,
                path: `${__dirname}/src/images`,
            },
        },
        {
            resolve: `gatsby-source-filesystem`,
            options: {
                name: `posts`,
                path: `${__dirname}/src/content/posts`,
            },
        },
        {
            resolve: `gatsby-source-filesystem`,
            options: {
                name: `pages`,
                path: `${__dirname}/src/content/pages`,
            },
        },
        `gatsby-transformer-remark`,
        {
            resolve: `gatsby-plugin-robots-txt`,
            options: {
                configFile: `robots-txt.config.js`,
            },
        },
        `gatsby-transformer-sharp`,
        `gatsby-plugin-sharp`,
        {
            resolve: `gatsby-plugin-manifest`,
            options: {
                name: `codeselfstudy`,
                short_name: `codeselfstudy`,
                start_url: `/`,
                background_color: `#666666`,
                theme_color: `#666666`,
                display: `minimal-ui`,
                icon: `src/images/favicon.png`, // This path is relative to the root of the site.
            },
        },
        // this (optional) plugin enables Progressive Web App + Offline functionality
        // To learn more, visit: https://gatsby.dev/offline
        // `gatsby-plugin-offline`,
        {
            resolve: `gatsby-plugin-purgecss`,
            options: {
                printRejected: true, // Print removed selectors and processed file names
                develop: false, // Enable while using `gatsby develop`
                // whitelist: ['whitelist'], // Don't remove this selector
                // ignore: ['/ignored.css', 'prismjs/', 'docsearch.js/'], // Ignore files/folders
                // purgeOnly : ['components/', '/main.css', 'bootstrap/'], // Purge only these files/folders
            },
        },
    ],
};
