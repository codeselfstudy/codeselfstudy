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
    const PostTemplate = path.resolve("src/templates/PostTemplate.js");
    const PageTemplate = path.resolve("src/templates/PageTemplate.js");

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

    // create blog posts
    // You can link to prev and next in the templates:
    // ```jsx
    // <Link to={next.frontmatter.path} rel="next">
    //     {next.frontmatter.title}
    // </Link>
    // ```
    each(postEdges, (edge, index) => {
        const prev =
            index === postEdges.length - 1 ? null : postEdges[index + 1].node;
        const next = index === 0 ? null : postEdges[index - 1].node;

        createPage({
            path: edge.node.frontmatter.path,
            component: PostTemplate,
            context: {
                slug: edge.node.frontmatter.path,
                prev,
                next,
            },
        });
    });

    // create pages
    each(pageEdges, (edge, _index) => {
        createPage({
            path: edge.node.frontmatter.path,
            component: PageTemplate,
            context: {
                slug: edge.node.frontmatter.path,
            },
        });
    });
};
