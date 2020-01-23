import React from "react";
import { Route, Switch } from "react-router-dom";
import Index from "../pages/index";
import About from "../pages/about";
import Puzzles from "../pages/puzzles";
import Puzzle from "../pages/puzzle";

const Routes = () => (
    <Switch>
        <Route exact path="/" component={Index} />
        <Route path="/about" component={About} />
        <Route path="/puzzles/:id" children={<Puzzle />} />
        <Route path="/puzzles" component={Puzzles} />
    </Switch>
);

export default Routes;
