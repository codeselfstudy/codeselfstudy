/**
 * Layout component that queries for data
 * with Gatsby's useStaticQuery component
 *
 * See: https://www.gatsbyjs.org/docs/use-static-query/
 */

import React from "react";

import Footer from "./Footer";
import Navbar from "./Navbar";
import Quotes from "./Quotes";
import SponsorsBox from "./SponsorsBox";
import "../styles/main.scss";

const Layout = ({ children, isHome = false }) => {
    return (
        <>
            <Navbar />
            <main>{children}</main>
            {isHome ? <SponsorsBox /> : null}
            <Quotes />
            <Footer />
        </>
    );
};

export default Layout;
