import React, { useState, useEffect } from "react";

import Layout from "../../../components/Layout";
import SEO from "../../../components/SEO";
import StepCheckbox from "./StepCheckbox";

import "./styles.scss";
// TODO: this is just a sketch

const formattingTools = [
    {
        language: "JavaScript",
        tools: [
            {
                name: "Prettier.js",
                url: "https://prettier.io/",
            },
        ],
    },
    {
        language: "Python",
        tools: [
            {
                name: "Flake8",
                url: "https://flake8.pycqa.org/en/latest/",
            },
        ],
    },
    {
        language: "Rust",
        tools: [
            {
                name: "rustfmt",
                url: "https://github.com/rust-lang/rustfmt",
            },
        ],
    },
    {
        language: "Elixir",
        tools: [
            {
                name: "mix format",
                url: "https://hexdocs.pm/mix/1.6.0/Mix.Tasks.Format.html",
            },
        ],
    },
    // add more here
];

const RubberDuck = () => {
    function handleSubmit(e) {
        e.preventDefault();
        alert("submitted [simulation]");
    }

    return (
        <Layout>
            <SEO title="Rubber Duck Debugging Tool - ask programming questions" />
            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">Rubber Duck Debugging Tool</h1>
                    <form>
                        <div className="step">
                            <div className="instructions">
                                <h2 className="title is-2">
                                    Step 1: Format Your Code Perfectly
                                </h2>
                                <div>
                                    <p>
                                        Make sure that your code is pefectly
                                        formatted and that the indentation is
                                        correct. This will help you find errors.
                                    </p>
                                    <button className="button">
                                        I don't know how to format my code
                                    </button>
                                    <p>
                                        That button can open up the table below
                                        in an accordion.
                                    </p>
                                    <h3 className="title is-3">
                                        Formatting Tips
                                    </h3>
                                    <p>
                                        You can integrate the tools below into
                                        your editor to automatically format your
                                        code when you save it. If you don't know
                                        how to integrate them into your editor,
                                        search online for the name of your
                                        editor and the name of the tool.
                                        Example:{" "}
                                        <a
                                            href="https://duckduckgo.com/?q=vscode+prettier"
                                            target="_blank"
                                            rel="noopener, noreferrer"
                                        >
                                            vscode prettier
                                        </a>
                                        .
                                    </p>
                                    <table className="table is-striped">
                                        <thead>
                                            <tr>
                                                <th>Language</th>
                                                <th>Tool(s)</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {formattingTools.map(t => (
                                                <tr>
                                                    <td>{t.language}</td>
                                                    <td>
                                                        {t.tools.map(i => (
                                                            <a
                                                                href={i.url}
                                                                target="_blank"
                                                                rel="noopener, noreferrer"
                                                            >
                                                                {i.name}
                                                            </a>
                                                        ))}
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            </div>

                            <StepCheckbox
                                step="1"
                                label="I have formatted my code with a formatting tool and still don't see the error."
                            />
                        </div>

                        <div className="step">
                            <div className="instructions">
                                <h2 className="title is-2">
                                    Step 2: Use a Code Linter
                                </h2>
                                <p>
                                    Run your code through a linter to check for
                                    code errors.
                                </p>
                                <StepCheckbox
                                    step="2"
                                    label="I have run my code through a linter and there are no errors."
                                />
                            </div>
                        </div>

                        <div className="step">
                            <div className="instructions">
                                <h2 className="title is-2">
                                    Step 3: Search Online for the Error Message
                                </h2>

                                <p>
                                    Use a search engine to search for the text
                                    of the error message.
                                </p>
                                <p>
                                    [[maybe they should include at least three
                                    URLs or resources they have read while
                                    trying to solve the problem, along with
                                    exactly why those solutions didn't work.]]
                                </p>
                                <StepCheckbox
                                    step="3"
                                    label="I have searched online for the error messages."
                                />
                            </div>
                        </div>

                        <div className="step">
                            <div className="instructions">
                                <h2 className="title is-2">
                                    Step 4: Talk to the Rubber Duck
                                </h2>

                                <p>
                                    Explain what each line and word in the code
                                    means to the rubber duck. Most errors can be
                                    solved this way.
                                </p>

                                <StepCheckbox
                                    step="4"
                                    label="I have explained my problem to the rubber duck"
                                />
                            </div>
                        </div>

                        <div>
                            If that still doesn’t fix the problem, then ask the
                            question in a form. We need more information than
                            “my code didn’t work.”
                        </div>

                        <div>
                            What did you expect to see happen in your code?
                            [textarea]
                        </div>
                        <div>What did you actually see? [textarea]</div>
                        <div>What is the full error message? [textarea]</div>
                        <div>What have you tried so far? [textarea]</div>
                        <div>
                            Post a link to the full, public source code (Github,
                            Gitlab, Bitbucket, etc.) If it’s a private repo,
                            upload a zip file of your repo into Slack or be
                            ready to invite other users to your private repo
                            [text input].
                        </div>
                        <div></div>
                        <div>
                            After all the checkboxes and fields are filled out,
                            it could auto-post to the forum and then send a link
                            to the forum post into Slack.
                        </div>
                    </form>
                </div>
            </section>
        </Layout>
    );
};

export default RubberDuck;
