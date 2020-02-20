import React from "react";
import { Link } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";
import "../styles/blog.scss";

function PostLink({ post }) {
    return (
        <div className="post-link box">
            <h2 className="title is-2">
                <Link to={post.frontmatter.path}>{post.frontmatter.title}</Link>
            </h2>
            <div>{post.excerpt}</div>
        </div>
    );
}

export default function Blog({
    data: {
        allMarkdownRemark: { edges },
    },
}) {
    const posts = edges
        .filter(edge => !!edge.node.frontmatter.date)
        .filter(edge => edge.node.fields.collection === "posts")
        .map(edge => <PostLink key={edge.node.id} post={edge.node} />);
    return (
        <Layout>
            <SEO title="Blog" />
            <section className="section blog-list">
                <div className="container content">
                    <h1 className="title is-1">Blog</h1>
                    <div>{posts}</div>
                </div>
            </section>
        </Layout>
    );
}
export const pageQuery = graphql`
    query {
        allMarkdownRemark(sort: { order: DESC, fields: [frontmatter___date] }) {
            edges {
                node {
                    id
                    excerpt(pruneLength: 250)
                    fields {
                        collection
                    }
                    frontmatter {
                        date(formatString: "MMMM DD, YYYY")
                        path
                        title
                    }
                }
            }
        }
    }
`;
