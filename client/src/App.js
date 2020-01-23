import React from "react";
import { BrowserRouter as Router, Switch, Route, Link } from "react-router-dom";

import Index from "./pages/index";
import About from "./pages/about";
import Puzzles from "./pages/puzzles";
import Puzzle from "./pages/puzzle";

function App() {
    return (
        <Router>
            <nav className="navbar navbar-expand-md navbar-dark bg-dark mb-4">
                <Link className="navbar-brand" to="/">
                    Coding Puzzles
                </Link>
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
                        <li className="nav-item active">
                            <Link className="nav-link" to="/about">
                                About
                            </Link>
                        </li>
                        <li className="nav-item">
                            <Link className="nav-link" to="/puzzles">
                                Puzzles
                            </Link>
                        </li>
                    </ul>
                </div>
            </nav>

            <Switch>
                <Route path="/about">
                    <About />
                </Route>

                <Route path="/puzzles/:id" children={<Puzzle />} />
                <Route path="/puzzles">
                    <Puzzles />
                </Route>
                <Route path="/">
                    <Index />
                </Route>
            </Switch>
        </Router>
    );
}

export default App;
