/**
 * Layout component that queries for data
 * with Gatsby's useStaticQuery component
 *
 * See: https://www.gatsbyjs.org/docs/use-static-query/
 */

import React from "react";
import { Helmet } from "react-helmet";

import Footer from "./Footer";
import FooterAnalytics from "./FooterAnalytics";
import Navbar from "./Navbar";
import Quotes from "./Quotes";
import SponsorsBox from "./SponsorsBox";
import "../styles/main.scss";

const Layout = ({ children, isHome = false }) => {
    return (
        <>
            <Helmet>
                <script>
                    var clicky_site_ids = clicky_site_ids || [];
                    clicky_site_ids.push(101016406);
                </script>
                <script async src="//static.getclicky.com/js"></script>
            </Helmet>
            <Navbar />
            <main>{children}</main>
            {isHome ? <SponsorsBox /> : null}
            <Quotes />
            <Footer />
            <FooterAnalytics />
        </>
    );
};

export default Layout;
