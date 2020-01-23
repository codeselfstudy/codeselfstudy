import React from "react";
import { useParams } from "react-router-dom";

export default function Puzzle(props) {
    const { id } = useParams();
    return (
        <div>
            <h1>Coding Puzzle #{id}</h1>
            <p>puzzle description goes here</p>
        </div>
    );
}
