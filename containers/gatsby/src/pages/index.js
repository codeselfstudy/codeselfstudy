import React from "react";
import { graphql } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

export default function IndexPage({ data }) {
    const page = data.allIndexToml.edges.filter(
        item => item.node.is_home === true
    )[0].node;

    return (
        <Layout>
            <SEO isHome={true} title={page.html_title} />

            <div dangerouslySetInnerHTML={{ __html: page.body }}></div>
        </Layout>
    );
}

// `data.allIndexToml.edges.node`
export const query = graphql`
    query HomeQuery {
        allIndexToml {
            edges {
                node {
                    is_home
                    title
                    html_title
                    body
                }
            }
        }
    }
`;
