import React from "react";

export default function Index() {
    return (
        <div>
            <h1 className="text-4xl">Home Page</h1>
            <p>Coding puzzles.</p>
            <div className="flex mb-4">
                <div className="w-1/3 bg-gray-400 h-12">
                    <h2>Hello</h2>
                </div>
                <div className="w-1/3 bg-gray-500 h-12">
                    <h2>Hello</h2>
                </div>
                <div className="w-1/3 bg-gray-400 h-12">
                    <h2>Hello</h2>
                </div>
            </div>
        </div>
    );
}
