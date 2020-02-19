import { Link } from "gatsby";
import React from "react";

export default function Footer({ siteTitle }) {
    return (
        <footer className="footer page-footer">
            <div className="container content">
                <div className="columns">
                    <div className="column">
                        <Link to="/">Code Self Study</Link> &bull; Programming
                        community in Berkeley, California
                    </div>
                    <div className="column">
                        <ul>
                            <li>
                                <Link to="/contact/">Contact Us</Link>
                            </li>
                            <li>
                                <Link to="/learn/">Learn How to Code</Link>
                            </li>
                            <li>
                                <Link to="/sponsors/">Sponsors</Link>
                            </li>
                            <li>
                                <Link to="/puzzles/">Coding Puzzles</Link>
                            </li>
                        </ul>
                        {/*
                        <ul>
                            <li>
                                <Link to="/about/">About</Link>
                            </li>
                            <li>
                                <Link to="/events/">Events</Link>
                            </li>
                            <li>
                                <Link to="/school/">School</Link>
                            </li>
                            <li>
                                <Link to="/jobs/">Jobs</Link>
                            </li>
                            <li>
                                <Link to="/hire-programmers/">
                                    Hire Programmers
                                </Link>
                            </li>
                            <li>
                                <Link to="/autism/">Autism Resources</Link>
                            </li>
                            <li>
                                <Link to="/contact/">Contact</Link>
                            </li>
                        </ul>
                        */}
                    </div>
                </div>
            </div>
        </footer>
    );
}
