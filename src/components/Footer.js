import { Link } from "gatsby";
import React from "react";
import { FaGithub, FaYoutube, FaTwitter } from "react-icons/fa";
import "../styles/footer.scss";

export default function Footer({ siteTitle }) {
    return (
        <footer className="footer page-footer">
            <div className="container content">
                <div className="columns">
                    <div className="column">
                        <ul>
                            <li>
                                <Link to="/contact/">Contact Us</Link>
                            </li>
                            <li>
                                <Link to="/learn/">Learn How to Code</Link>
                            </li>
                            <li>
                                <Link to="/puzzles/">Coding Puzzles</Link>
                            </li>
                            <li>
                                <a href="https://blog.codeselfstudy.com/">
                                    Blog
                                </a>
                            </li>
                            <li>
                                <Link to="/school/">School</Link>
                            </li>
                        </ul>
                    </div>
                    <div className="column">
                        <ul>
                            <li>
                                <Link to="/jobs/">Jobs</Link>
                            </li>
                            <li>
                                <Link to="/about/">About</Link>
                            </li>
                            <li>
                                <Link to="/events/">Events</Link>
                            </li>
                            <li>
                                <Link to="/autism/">Autism Resources</Link>
                            </li>
                            <li>
                                <a href="https://forum.codeselfstudy.com/">
                                    Forum
                                </a>
                            </li>
                        </ul>
                    </div>
                    <div className="column icons">
                        <a href="https://github.com/codeselfstudy/codeselfstudy">
                            <FaGithub title="Source code on Github" size={32} />
                        </a>
                        <br />
                        <a href="https://twitter.com/codeselfstudy">
                            <FaTwitter
                                title="@codeselfstudy on Twitter"
                                size={32}
                            />
                        </a>
                        <br />
                        <a href="https://www.youtube.com/codeselfstudy">
                            <FaYoutube
                                title="Code Self Study on YouTube"
                                size={32}
                            />
                        </a>
                        <br />
                    </div>
                </div>
                <div>
                    <p>
                        <Link to="/">Code Self Study</Link> &bull; Programming
                        community in Berkeley &amp; San Francsico Bay Area,
                        California
                    </p>
                    <p>
                        Hosted on{" "}
                        <Link
                            to="/discount/digital-ocean/"
                            className="digital-ocean"
                        >
                            DigitalOcean
                        </Link>
                    </p>
                </div>
            </div>
        </footer>
    );
}
