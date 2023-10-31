import React from "react";

function Topic({ data }) {
    const { title, slug, id } = data;

    const url = `https://forum.codeselfstudy.com/t/${slug}/${id}`;
    return (
        <div
            style={{
                margin: "17px 0",
            }}
        >
            <a
                style={{ textDecoration: "none" }}
                href={url}
                target="_blank"
                rel="noopener noreferrer"
            >
                {title}
            </a>
        </div>
    );
}

export default function Topics({ topics, offset = 0, numTopics = 14 }) {
    const halfway = Math.floor(numTopics / 2);
    const leftTopics = topics.slice(offset, halfway + offset);
    const rightTopics = topics.slice(halfway + offset, numTopics + offset);

    return (
        <>
            <div className="column">
                {leftTopics.map((t) => (
                    <Topic data={t} />
                ))}
            </div>
            <div className="column">
                {rightTopics.map((t) => (
                    <Topic data={t} />
                ))}
            </div>
        </>
    );
}
