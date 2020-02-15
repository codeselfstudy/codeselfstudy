import React from "react";
import { Link } from "gatsby";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

export default function NotFoundPage() {
    return (
        <Layout>
            <SEO title="404: Not found" />
            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">NOT FOUND</h1>
                    <p>
                        Some things on the website were moved around recently,
                        and the page you requested wasn't found.
                    </p>
                    <p>
                        Please <Link to="/contact/">contact us</Link> if you've
                        discovered a problem, or see if{" "}
                        <Link to="/">the home page</Link> links to the
                        information you're looking for.
                    </p>
                </div>
            </section>
        </Layout>
    );
}
