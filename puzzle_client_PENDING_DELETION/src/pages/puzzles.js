import React from "react";
import { Link } from "react-router-dom";

export default function Puzzles() {
    return (
        <div className="container">
            <h1>Coding Puzzles</h1>
            <p>Dynamic routes:</p>
            <ul>
                <li>
                    <Link className="btn btn-lg btn-primary" to="/puzzles/123">
                        Puzzle 123
                    </Link>
                </li>
            </ul>
        </div>
    );
}
