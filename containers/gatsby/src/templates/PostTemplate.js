import React from "react";
import { graphql, Link } from "gatsby";

import Layout from "../components/Layout";

const metadataStyle = {
    color: "#888",
    textTransform: "uppercase",
    fontSize: "0.8rem",
};

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
            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">{frontmatter.title}</h1>
                    <div className="blog-post-metadata" style={metadataStyle}>
                        Posted by {frontmatter.author || "Admin"} on{" "}
                        {frontmatter.date}
                    </div>
                    <div
                        className="blog-post-content"
                        dangerouslySetInnerHTML={{ __html: html }}
                    />
                    <hr />
                    <div>
                        <Link
                            to="/blog/"
                            className="button is-large is-default"
                        >
                            Return to the Blog
                        </Link>
                    </div>
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
                title
                author
            }
        }
    }
`;
