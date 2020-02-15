/**
 * Implement Gatsby's Node APIs in this file.
 *
 * See: https://www.gatsbyjs.org/docs/node-apis/
 */

const path = require("path");
const { get, each } = require("lodash");

exports.onCreateNode = ({ node, actions, getNode }) => {
    const { createNodeField } = actions;

    if (get(node, "internal.type") === "MarkdownRemark") {
        const parent = getNode(get(node, "parent"));
        createNodeField({
            node,
            name: "collection",
            value: get(parent, "sourceInstanceName"),
        });
    }
};

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
                        fields {
                            collection
                        }
                        frontmatter {
                            path
                            title
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
    const postEdges = allEdges.filter(edge => {
        return edge.node.fields.collection === "posts";
    });
    const pageEdges = allEdges.filter(edge => {
        return edge.node.fields.collection === "pages";
    });

    console.log(
        "postEdges",
        postEdges.map(p => JSON.stringify(p.node))
    );
    console.log(
        "pageEdges",
        pageEdges.map(p => JSON.stringify(p.node))
    );

    // | postEdges [
    // |   '{"fields":{"collection":"posts"},"frontmatter":{"path":"/blog/title":"Test Post"}}'
    // | ]
    // | pageEdges [
    // |   '{"fields":{"collection":"pages"},"frontmatter":{"path":"/about/","title":"About"}}'
    // | ]

    // | postEdges [
    // |   [Object: null prototype] {
    // |     node: [Object: null prototype] {
    // |       fields: [Object: null prototype],
    // |       frontmatter: [Object: null prototype]
    // |     }
    // |   }
    // | ]
    // | pageEdges [
    // |   [Object: null prototype] {
    // |     node: [Object: null prototype] {
    // |       fields: [Object: null prototype],
    // |       frontmatter: [Object: null prototype]
    // |     }
    // |   }
    // | ]

    postEdges.forEach(({ node }) => {
        createPage({
            path: node.frontmatter.path,
            component: postTemplate,
            context: {}, // additional data can be passed via context
        });
    });

    // pageEdges.forEach(({ node }) => {
    //     createPage({
    //         path: node.frontmatter.path,
    //         component: pageTemplate,
    //         context: {}, // additional data can be passed via context
    //     });
    // });
};
