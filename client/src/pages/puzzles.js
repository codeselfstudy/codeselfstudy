import React from "react";
import { Link } from "react-router-dom";

export default function Puzzles() {
    return (
        <div>
            <h1>Coding Puzzles</h1>
            <p>Dynamic routes:</p>
            <ul>
                <li>
                    <Link to="/puzzles/123">Puzzle 123</Link>
                </li>
            </ul>
        </div>
    );
}
