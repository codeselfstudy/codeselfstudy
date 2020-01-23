import React from "react";
import { BrowserRouter as Router, Switch, Route, Link } from "react-router-dom";

import Index from "./pages/index";
import About from "./pages/about";
import Puzzles from "./pages/puzzles";
import Puzzle from "./pages/puzzle";

function App() {
    return (
        <Router>
            <nav>
                <ul className="flex border-b">
                    <li className="-mb-px mr-1">
                        <Link
                            className="bg-white inline-block border-l border-t border-r rounded-t py-2 px-4 text-blue-700 font-semibold"
                            to="/"
                        >
                            Home
                        </Link>
                    </li>
                    <li className="mr-1">
                        <Link
                            className="bg-white inline-block py-2 px-4 text-blue-500 hover:text-blue-800 font-semibold"
                            to="/about"
                        >
                            About
                        </Link>
                    </li>
                    <li className="mr-1">
                        <Link
                            className="bg-white inline-block py-2 px-4 text-blue-500 hover:text-blue-800 font-semibold"
                            to="/puzzles"
                        >
                            Puzzles
                        </Link>
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
