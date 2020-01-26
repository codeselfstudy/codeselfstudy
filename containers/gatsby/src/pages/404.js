import React from "react";

import Layout from "../components/Layout";
import SEO from "../components/SEO";

export default function NotFoundPage() {
    return (
        <Layout>
            <SEO title="404: Not found" />
            <h1>NOT FOUND</h1>
            <p>Page not found.</p>
        </Layout>
    );
}
