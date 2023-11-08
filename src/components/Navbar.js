import { Link } from "gatsby";
import React, { useState } from "react";

const Navbar = () => {
    const [navIsOpen, setNavIsOpen] = useState(false);

    const toggleNav = () => {
        setNavIsOpen(!navIsOpen);
    };

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
                        className={`navbar-burger burger ${
                            navIsOpen ? "is-active" : ""
                        }`}
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

                <div
                    id="codeselfstudyNavbar"
                    className={`navbar-menu ${navIsOpen ? "is-active" : ""}`}
                >
                    <div className="navbar-start">
                        <Link to="/about/" className="navbar-item">
                            About
                        </Link>

                        <Link to="/events/" className="navbar-item">
                            Events
                        </Link>

                        {/* <a
                            href="https://forum.codeselfstudy.com/"
                            className="navbar-item"
                        >
                            Forum
                        </a> */}
                    </div>

                    {/* <div className="navbar-end">
                        <div className="navbar-item">
                            <div className="buttons">
                                <a
                                    className="button is-primary"
                                    href="https://forum.codeselfstudy.com/signup"
                                    rel="nofollow"
                                >
                                    <strong>Sign Up</strong>
                                </a>
                                <a
                                    className="button is-light"
                                    href="https://forum.codeselfstudy.com/login"
                                    rel="nofollow"
                                >
                                    Log In
                                </a>
                            </div>
                        </div>
                    </div> */}
                </div>
            </div>
        </nav>
    );
};
export default Navbar;
