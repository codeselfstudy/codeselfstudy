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
        language: "C/C++",
        tools: [
            {
                name: "ClangFormat",
                url: "https://clang.llvm.org/docs/ClangFormat.html",
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
                <div className="container">
                    <ul class="steps is-thin">
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment is-active">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                        <li class="steps-segment">
                            <span class="steps-marker"></span>
                        </li>
                    </ul>
                    <div className="content">
                        <h1 className="title is-1">
                            Rubber Duck Debugging Tool
                        </h1>
                        <form onSubmit={handleSubmit}>
                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 1: Format Your Code Perfectly
                                    </h2>
                                    <div>
                                        <p>
                                            Make sure that your code is pefectly
                                            formatted and that the indentation
                                            is correct. This will help you find
                                            errors.
                                        </p>
                                        <button className="button is-warning">
                                            I don't know how to format my code
                                        </button>
                                        <p>
                                            That button can open up the table
                                            below in an accordion.
                                        </p>
                                        <div style={{ opacity: "0.5" }}>
                                            <h3 className="title is-3">
                                                Formatting Tips
                                            </h3>
                                            <p>
                                                If you don't know how to
                                                automatically format your code,
                                                you can do it manually. Or ask
                                                us in{" "}
                                                <a
                                                    href="https://forum.codeselfstudy.com/"
                                                    target="_blank"
                                                    rel="noopener, noreferrer"
                                                >
                                                    the forum
                                                </a>
                                                , and we'll be happy to help.
                                                This step is required, because
                                                if you don't format your own
                                                code, the person helping you is
                                                going to have to format it, and
                                                if it turns out to be a problem
                                                that could have been solved by
                                                autoformatting your code.
                                            </p>
                                            <p>
                                                You can integrate the tools
                                                below into your editor to
                                                automatically format your code
                                                when you save it. If you don't
                                                know how to integrate them into
                                                your editor, search online for
                                                the name of your editor and the
                                                name of the tool. Example:{" "}
                                                <a
                                                    href="https://duckduckgo.com/?q=vscode+prettier"
                                                    target="_blank"
                                                    rel="noopener, noreferrer"
                                                >
                                                    vscode prettier
                                                </a>
                                                . If you experience errors while
                                                trying to install the code
                                                formatter, you can use the
                                                checklist on this page to solve
                                                the problem.
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
                                                            <td>
                                                                {t.language}
                                                            </td>
                                                            <td>
                                                                {t.tools.map(
                                                                    i => (
                                                                        <a
                                                                            href={
                                                                                i.url
                                                                            }
                                                                            target="_blank"
                                                                            rel="noopener, noreferrer"
                                                                        >
                                                                            {
                                                                                i.name
                                                                            }
                                                                        </a>
                                                                    )
                                                                )}
                                                            </td>
                                                        </tr>
                                                    ))}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                </div>

                                <StepCheckbox
                                    step="1"
                                    label="I have formatted my code well and still don't see the error."
                                />
                            </div>

                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 2: Use a Code Linter
                                    </h2>
                                    <p>
                                        Run your code through a linter to check
                                        for code errors. This is often built
                                        into editors.
                                    </p>
                                    <p>
                                        [[Should a button go here with
                                        instructions for viewing errors in
                                        vscode?]]
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
                                        Step 3: Search Online for the Error
                                        Message
                                    </h2>

                                    <p>
                                        Use a search engine to search for the
                                        text of the error message.
                                    </p>
                                    <p>
                                        [[maybe they should include at least
                                        three URLs or resources they have read
                                        while trying to solve the problem, along
                                        with exactly why those solutions didn't
                                        work.]]
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
                                        Explain what each line and word in the
                                        code means to the rubber duck. Most
                                        errors can be solved this way.
                                    </p>

                                    <StepCheckbox
                                        step="4"
                                        label="I have explained my problem to the rubber duck"
                                    />
                                </div>
                            </div>

                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 5: Start Asking Your Question
                                    </h2>
                                    <div>
                                        If that still doesn’t fix the problem,
                                        then ask the question in a form. We need
                                        more information than “my code didn’t
                                        work.”
                                    </div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Please be detailed"
                                    ></textarea>
                                    <StepCheckbox
                                        step="5"
                                        label="I have explained my problem in detail in the textarea above."
                                    />
                                </div>
                            </div>

                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 6: What Did You Expect to See?
                                    </h2>
                                    <div>
                                        What did you expect to see happen in
                                        your code?
                                    </div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Please be detailed"
                                    ></textarea>
                                    <StepCheckbox
                                        step="5"
                                        label="I have explained what I expected to see in the textarea above."
                                    />
                                </div>
                            </div>
                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 7: What Did You Actually See?
                                    </h2>
                                    <div>What did you actually see?</div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Please be detailed"
                                    ></textarea>
                                    <StepCheckbox
                                        step="4"
                                        label="I have explained my problem to the rubber duck"
                                    />
                                </div>
                            </div>
                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 8: What Is the Full Error Message?
                                    </h2>
                                    <div>
                                        What is the full error message? Be sure
                                        to remove any private details like
                                        secret keys, passwords, or other
                                        sensitive information.
                                    </div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Don't paste any passwords here."
                                    ></textarea>
                                    <StepCheckbox
                                        step="8"
                                        label="I have pasted my error messages in the textarea above."
                                    />
                                </div>
                            </div>
                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 9: What Have You Tried So Far?
                                    </h2>
                                    <div>
                                        What have you tried so far? What
                                        happened when you tried those things?
                                    </div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Please be detailed"
                                    ></textarea>
                                    <StepCheckbox
                                        step="9"
                                        label="I have explained what I've tried so far."
                                    />
                                </div>
                            </div>
                            <div className="step">
                                <div className="instructions">
                                    <h2 className="title is-2">
                                        Step 10: Show Us the Code
                                    </h2>
                                    <div>
                                        Post a link to the full, public source
                                        code (Github, Gitlab, Bitbucket, etc.)
                                        If it’s a private repo, upload a zip
                                        file of your repo into Slack or be ready
                                        to invite other users to your private
                                        repo.
                                    </div>
                                    <textarea
                                        className="textarea"
                                        rows="10"
                                        placeholder="Link to the code"
                                    ></textarea>
                                    <StepCheckbox
                                        step="10"
                                        label="I have linked to the code or explained why the code isn't available."
                                    />
                                </div>
                            </div>

                            <div className="step">
                                <h2 className="title is-2">Ask for Help</h2>
                                <p className="subtitle">
                                    You've made it to the end of the form and
                                    are ready to ask for help!
                                </p>
                                <p>
                                    Click the button below to submit the
                                    question to a to a group of people who can
                                    help. The content above will be posted to{" "}
                                    <a
                                        href="https://forum.codeselfstudy.com/"
                                        target="_blank"
                                        rel="noopener, noreferrer"
                                    >
                                        the forum
                                    </a>{" "}
                                    and you should receive a reply soon!
                                </p>
                                <button className="button is-danger is-large">
                                    SUMMON HELP
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </section>
        </Layout>
    );
};

export default RubberDuck;
