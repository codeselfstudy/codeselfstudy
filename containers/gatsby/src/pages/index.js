import React, { useEffect, useState } from "react";
import { Link } from "gatsby";
import axios from "axios";

import Layout from "../components/Layout";
import SEO from "../components/SEO";
import Spinner from "../components/Spinner";

export default function IndexPage() {
    const [isLoading, setIsLoading] = useState(true);
    const [apiData, setApiData] = useState(null);

    useEffect(() => {
        axios.get(`/api/puzzles/123`).then(res => {
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
                {isLoading ? (
                    <Spinner />
                ) : (
                    <code>{JSON.stringify(apiData)}</code>
                )}
                <p>
                    There's also a <Link to="/page-2/">page 2</Link>.
                </p>
            </section>
        </Layout>
    );
}
