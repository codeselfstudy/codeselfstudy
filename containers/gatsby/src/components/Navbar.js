import React, { useState, useEffect } from "react";
import { Link } from "gatsby";

const Navbar = () => {
    const toggleNav = () => {
        const toggle = document.querySelector(".navbar-burger");
        const menu = document.querySelector(".navbar-menu");
        toggle.classList.toggle("is-active");
        menu.classList.toggle("is-active");
    };

    const [currentUrl, setCurrentUrl] = useState("https://codeselfstudy.com/");
    useEffect(() => {
        // `window` doesn't exist in Gatsby during the Node.js build process.
        if (typeof window !== "undefined") {
            setCurrentUrl(encodeURI(window.location.href));
        }
    }, []);

    return (
        <nav
            id="mainNavbar"
            className="navbar is-fixed-top"
            role="navigation"
            aria-label="main navigation"
        >
            <div className="container">
                <div className="navbar-brand">
                    <Link to="/" className="navbar-item" id="siteLogoText">
                        Code Self Study
                    </Link>

                    {/* eslint-disable-next-line */}
                    <a
                        role="button"
                        className="navbar-burger burger"
                        onClick={toggleNav}
                        aria-label="menu"
                        aria-expanded="false"
                        data-target="codeselfstudyNavbar"
                    >
                        <span aria-hidden="true"></span>
                        <span aria-hidden="true"></span>
                        <span aria-hidden="true"></span>
                    </a>
                </div>

                <div id="codeselfstudyNavbar" className="navbar-menu">
                    <div className="navbar-start">
                        <Link to="/about/" className="navbar-item">
                            About
                        </Link>

                        <Link to="/events/" className="navbar-item">
                            Events
                        </Link>

                        <a
                            href="https://forum.codeselfstudy.com/"
                            className="navbar-item"
                        >
                            Forum
                        </a>
                    </div>

                    <div className="navbar-end">
                        <div className="navbar-item">
                            <div className="buttons">
                                {/*
                                    The current thinking here is that
                                    people signing up via the top navbar
                                    can go to the forum, and people
                                    who are logging in probably want
                                    to return to the same page. We can
                                    discuss it in the Github issues.
                                */}
                                <a
                                    className="button is-primary"
                                    href="https://forum.codeselfstudy.com/signup"
                                    rel="nofollow"
                                >
                                    <strong>Sign Up</strong>
                                </a>
                                <a
                                    className="button is-light"
                                    href={`/api/auth/request?destination=${currentUrl}`}
                                    rel="nofollow"
                                >
                                    <strong>Log In</strong>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </nav>
    );
};
export default Navbar;
