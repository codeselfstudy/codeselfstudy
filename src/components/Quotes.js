import React, { useState, useEffect } from "react";

import { quotes } from "../data/quotes";

export default function Quotes(props) {
    const [quote, setQuote] = useState("");
    const [author, setAuthor] = useState("");
    const [backgroundStyle, setBackgroundStyle] = useState("");

    useEffect(() => {
        const bgs = [
            "bg-dice",
            "bg-temple",
            "bg-steel-beams",
            "bg-circuit-board",
            "bg-topography",
            "bg-graph-paper",
            "bg-pie-factory",
            "bg-hexagons",
            "bg-cogs",
            "bg-bevel-circle",
        ];
        const currentQuote = quotes[Math.floor(Math.random() * quotes.length)];
        const currentBg = Math.floor(Math.random() * bgs.length);
        setBackgroundStyle(`hero quote-hero ${bgs[currentBg]}`);
        setQuote(currentQuote.quote);
        setAuthor(currentQuote.author);
    }, []);

    return (
        <div className={backgroundStyle}>
            <div className="hero-body">
                <blockquote className="container content">
                    <p>
                        {quote}
                        <br />
                        &mdash; {author}
                    </p>
                </blockquote>
            </div>
        </div>
    );
}
