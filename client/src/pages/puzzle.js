import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";

export default function Puzzle() {
    const { id } = useParams();
    const [puzzleData, setPuzzleData] = useState("loading");

    useEffect(() => {
        axios.get(`/api/puzzles/${id}`).then(res => setPuzzleData(res));
    }, [id]);

    return (
        <div>
            <h1 className="text-4xl">Coding Puzzle #{id}</h1>
            <p>puzzle description goes here</p>
            <code>{JSON.stringify(puzzleData)}</code>
        </div>
    );
}
