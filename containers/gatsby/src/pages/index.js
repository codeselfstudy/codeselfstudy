import React from "react";

import Layout from "../components/Layout";
// import SEO from "../components/SEO";
// import * as toml from "toml";
// const fs = require("fs");
// import * as path from "path";

// console.log("FS", fs);

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
    return (
        <Layout>
            <section className="section">
                <h1 className="title is-1">Future Homepage</h1>
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
