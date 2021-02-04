import React from "react";
import { graphql } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

/**
 * This is the template for the website's basic content pages.
 */
export default function PageTemplate({
    data, // this prop will be injected by the GraphQL query below.
}) {
    const { markdownRemark } = data; // data.markdownRemark holds your post data
    const { frontmatter, html } = markdownRemark;

    return (
        <Layout>
            <SEO title={frontmatter.html_title || frontmatter.title} />
            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">{frontmatter.title}</h1>
                    <div
                        className="page-content"
                        dangerouslySetInnerHTML={{ __html: html }}
                    />
                </div>
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
                html_title
                title
            }
        }
    }
`;
