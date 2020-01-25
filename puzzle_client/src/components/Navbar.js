import React from "react";
import { NavLink } from "react-router-dom";

export default function Navbar() {
    return (
        <nav className="navbar navbar-expand-md navbar-dark bg-dark mb-4">
            <NavLink
                exact
                activeClassName="active"
                className="navbar-brand"
                to="/"
            >
                Coding Puzzles
            </NavLink>
            <button
                className="navbar-toggler"
                type="button"
                data-toggle="collapse"
                data-target="#navbarCollapse"
                aria-controls="navbarCollapse"
                aria-expanded="false"
                aria-label="Toggle navigation"
            >
                <span className="navbar-toggler-icon"></span>
            </button>
            <div className="collapse navbar-collapse" id="navbarCollapse">
                <ul className="navbar-nav mr-auto">
                    <li className="nav-item">
                        <NavLink
                            activeClassName="active"
                            className="nav-link"
                            to="/about"
                        >
                            About
                        </NavLink>
                    </li>
                    <li className="nav-item">
                        <NavLink
                            activeClassName="active"
                            className="nav-link"
                            to="/puzzles"
                        >
                            Puzzles
                        </NavLink>
                    </li>
                </ul>
            </div>
        </nav>
    );
}
