import React from "react";
import { graphql, Link } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

/**
 * This is the template for blog posts.
 */
export default function PostTemplate({
    data, // this prop will be injected by the GraphQL query below.
}) {
    const { markdownRemark } = data; // data.markdownRemark holds your post data
    const { frontmatter, html } = markdownRemark;

    return (
        <Layout>
            <SEO
                isHome={true}
                title="Learn Computer Programming, Python, JavaScript at Our Bootcamp Alternative"
            />
            <section className="section">
                <h1 className="title is-1">{frontmatter.title}</h1>
                <div className="blog-post-metadata">{frontmatter.date}</div>
                <div
                    className="blog-post-content"
                    dangerouslySetInnerHTML={{ __html: html }}
                />
            </section>
        </Layout>
    );
}

export const pageQuery = graphql`
    query($path: String!) {
        markdownRemark(frontmatter: { path: { eq: $path } }) {
            html
            frontmatter {
                date(formatString: "MMMM DD, YYYY")
                path
                title
            }
        }
    }
`;
