import React, { useEffect, useState } from "react";
import { Link } from "gatsby";
import axios from "axios";

import Layout from "../components/Layout";
import SEO from "../components/SEO";
import Spinner from "../components/Spinner";

// TODO: This component can be removed once the real content is connected to Gatsby.
function DemoTable({ data }) {
    const entries = Object.entries(data);

    return (
        <table
            style={{ border: "2px dotted #666" }}
            className="table is-responsive"
        >
            <caption style={{ backgroundColor: "yellow" }}>
                A sample web request to the backend.
            </caption>
            <thead>
                <tr>
                    <th>key</th>
                    <th>value</th>
                </tr>
            </thead>
            <tbody>
                {entries.map(i => (
                    <tr>
                        <td>
                            <strong>{i[0]}</strong>
                        </td>
                        <td>
                            <code>{JSON.stringify(i[1])}</code>
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}

export default function IndexPage() {
    const [isLoading, setIsLoading] = useState(true);
    const [apiData, setApiData] = useState(null);

    useEffect(() => {
        axios
            .get(`/api/puzzles/${Math.floor(Math.random() * 2000)}`)
            .then(res => {
                setTimeout(() => {
                    setApiData(res);
                    setIsLoading(false);
                }, 2000);
            });
    }, []);

    return (
        <Layout>
            <SEO
                isHome={true}
                title="Learn Computer Programming, Python, JavaScript at Our Bootcamp Alternative"
            />
            <section className="section">
                <h1 className="title is-1">Hello World</h1>
                <p>This should load some fake data from the Express API.</p>

                {isLoading ? <Spinner /> : <DemoTable data={apiData} />}

                <p>
                    There's also a <Link to="/page-2/">page 2</Link>.
                </p>
            </section>
        </Layout>
    );
}
