import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";
import Spinner from "../components/Spinner";

export default function Puzzle() {
    const { id } = useParams();
    const [puzzleData, setPuzzleData] = useState();
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        axios.get(`/api/puzzles/${id}`).then(res => {
            setTimeout(() => {
                setPuzzleData(res);
                setIsLoading(false);
            }, 500);
        });
    }, [id]);

    return (
        <div className="container">
            <h1 className="text-4xl">Coding Puzzle #{id}</h1>
            <p>puzzle description goes here</p>
            <code>
                {isLoading ? <Spinner /> : <></>}
                <pre>{JSON.stringify(puzzleData, null, 4)}</pre>
            </code>
        </div>
    );
}
