module.exports = {
    siteMetadata: {
        title: `Code Self Study`,
        description: `Programming community in Berkeley, California`,
        author: `@codeselfstudy`,
        siteUrl:
            process.env.NODE_ENV === "production"
                ? "https://codeselfstudy.com"
                : "https://localhost:8000",
    },
    plugins: [
        `gatsby-plugin-sass`,
        `gatsby-transformer-toml`,
        `gatsby-plugin-react-helmet`,
        {
            resolve: `gatsby-source-filesystem`,
            options: {
                name: `images`,
                path: `${__dirname}/src/images`,
            },
        },
        // TODO: there will probably be job listings later. Use this as an
        // example.
        // {
        //     resolve: `gatsby-source-filesystem`,
        //     options: {
        //         name: `posts`,
        //         path: `${__dirname}/src/content/posts`,
        //     },
        // },
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
                // The whitelist rules match any prefix (class, id, element).
                // https://purgecss.com/whitelisting.html#specific-selectors
                whitelist: [
                    "blockquote",
                    "hr",
                    "hero",
                    "homepage-hero",
                    "bg-circuit-board",
                    "hero-body",
                    "container",
                    "content",
                ],
                // ignore: ['/ignored.css', 'prismjs/', 'docsearch.js/'], // Ignore files/folders
                // purgeOnly : ['components/', '/main.css', 'bootstrap/'], // Purge only these files/folders
            },
        },
    ],
};
