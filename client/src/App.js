import React from "react";
import { BrowserRouter as Router, Switch, Route, Link } from "react-router-dom";

import Index from "./pages/index";
import About from "./pages/about";
import Puzzles from "./pages/puzzles";
import Puzzle from "./pages/puzzle";

import "./App.scss";

function App() {
    return (
        <Router>
            <nav>
                <ul>
                    <li>
                        <Link to="/">Home</Link>
                    </li>
                    <li>
                        <Link to="/about">About</Link>
                    </li>
                    <li>
                        <Link to="/puzzles">Puzzles</Link>
                    </li>
                </ul>
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
