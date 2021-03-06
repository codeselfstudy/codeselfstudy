// import React, { useEffect, useState } from "react";
import React from "react";
// import axios from "axios";

import Layout from "../components/Layout";
import SEO from "../components/SEO";
// import Spinner from "../components/Spinner";

// TODO: This component can be removed once the real content is connected to Gatsby.
// function DemoTable({ data }) {
//     const entries = Object.entries(data);

//     return (
//         <table
//             style={{ border: "2px dotted #666" }}
//             className="table is-responsive"
//         >
//             <caption style={{ backgroundColor: "yellow" }}>
//                 A sample web request to the backend.
//             </caption>
//             <thead>
//                 <tr>
//                     <th>key</th>
//                     <th>value</th>
//                 </tr>
//             </thead>
//             <tbody>
//                 {entries.map(i => (
//                     <tr key={i[0]}>
//                         <td>
//                             <strong>{i[0]}</strong>
//                         </td>
//                         <td>
//                             <code>{JSON.stringify(i[1])}</code>
//                         </td>
//                     </tr>
//                 ))}
//             </tbody>
//         </table>
//     );
// }

export default function PuzzlesPage() {
    return (
        <Layout>
            <SEO title="Coding Puzzles and Algorithm Practice for Whiteboard Interview Prep" />
            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">Puzzles</h1>

                    <p>A page of coding puzzles is coming soon!</p>
                </div>
            </section>
        </Layout>
    );
}

// Below is sample code to hit the Express API
// export default function PuzzlesPage() {
//     const [isLoading, setIsLoading] = useState(true);
//     const [apiData, setApiData] = useState(null);

//     useEffect(() => {
//         axios.get(`/api/puzzles`).then(res => {
//             console.log("res", res);
//             setTimeout(() => {
//                 setApiData(res.data.puzzles[0]);
//                 setIsLoading(false);
//             }, 1000);
//         });
//     }, []);

//     return (
//         <Layout>
//             <SEO title="Coding Puzzles and Algorithm Practice for Whiteboard Interview Prep" />
//             <section className="section">
//                 <h1 className="title is-1">Puzzles</h1>

//                 {isLoading ? <Spinner /> : <DemoTable data={apiData} />}
//             </section>
//         </Layout>
//     );
// }
