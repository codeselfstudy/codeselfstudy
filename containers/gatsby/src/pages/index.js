import React from "react";
import { graphql } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

// // The home page content is loaded from a TOML file (git submodule).
// const tomlFile = path.resolve("../content/pages/index.toml");
// console.log("========");
// console.log("tomlFile", tomlFile);
// const rawToml = fs.readFileSync(tomlFile);
// console.log("rawToml", rawToml);
// const pageData = toml.parse(rawToml);
// console.log("pageData", pageData);
// console.log("========");

export default function IndexPage() {
    // const { title, html_title, body } = data.allIndexToml.edges.node;
    return (
        <Layout>
            <SEO isHome={true} title="home" />
            <section className="section">
                <h1 className="title is-1">ASDF</h1>
                Dolor fugiat cumque amet eius distinctio enim? Quaerat provident
                praesentium quo nam qui, dolorum. A dolores molestiae
                praesentium veniam soluta Iste quia provident impedit quaerat
                quae? Quia ratione nostrum asperiores
            </section>
        </Layout>
    );
}

// return (
//     <Layout>
//         <SEO isHome={true} title={pageData.html_title} />
//         <section className="section">
//             <h1 className="title is-1">{pageData.title}</h1>
//             {pageData.body}
//         </section>
//     </Layout>
// );

// `data.allIndexToml.edges.node`
export const query = graphql`
    query HomeQuery {
        allIndexToml {
            edges {
                node {
                    title
                    html_title
                    body
                }
            }
        }
    }
`;
