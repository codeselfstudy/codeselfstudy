/**
 * Implement Gatsby's Node APIs in this file.
 *
 * See: https://www.gatsbyjs.org/docs/node-apis/
 */

const path = require("path");
const { get, each } = require("lodash");

// exports.onCreateNode = ({ node, actions, getNode }) => {
//     const { createNodeField } = actions;

//     if (get(node, "internal.type") === "MarkdownRemark") {
//         const parent = getNode(get(node, "parent"));
//         createNodeField({
//             node,
//             name: "collection",
//             value: get(parent, "sourceInstanceName"),
//         });
//     }
// };

exports.createPages = async ({ actions, graphql, reporter }) => {
    const { createPage } = actions;
    const postTemplate = path.resolve("src/templates/PostTemplate.js");
    const pageTemplate = path.resolve("src/templates/pageTemplate.js");

    const result = await graphql(`
        {
            allMarkdownRemark(
                sort: { order: DESC, fields: [frontmatter___date] }
                limit: 1000
            ) {
                edges {
                    node {
                        frontmatter {
                            path
                        }
                    }
                }
            }
        }
    `);
    // Handle errors
    if (result.errors) {
        reporter.panicOnBuild(`Error while running GraphQL query.`);
        return;
    }

    // Separate the collection types
    const allEdges = result.data.allMarkdownRemark.edges;
    // const postEdges = allEdges.filter(
    //     edge => edge.node.fields.collection === "posts"
    // );
    // const pageEdges = allEdges.filter(
    //     edge => edge.node.fields.collection === "pages"
    // );

    result.data.allMarkdownRemark.edges.forEach(({ node }) => {
        createPage({
            path: node.frontmatter.path,
            component: postTemplate,
            context: {}, // additional data can be passed via context
        });
    });
};
