import React, { useEffect, useState } from "react";
import Spinner from "react-text-spinners";
import wretch from "wretch";

import "./ForumTopics.scss";
import Topics from "./Topics";

export default function ForumPosts(props) {
    const [topics, setTopics] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        topicFetcher();
    }, []);

    async function topicFetcher() {
        const base_url = "https://forum.codeselfstudy.com";
        await wretch(`${base_url}/latest.json`)
            .get()
            .json(data => {
                const ts = data["topic_list"]["topics"];
                console.log("got topics", ts);
                setTopics(ts);
            })
            .catch(err => {
                console.error("ERROR", err);
                setError(`${err.name}: ${err.message}`);
            });
    }

    return (
        <div className="forum-posts">
            <section className="section">
                <div className="container content">
                    <div className="box">
                        <h2 className="title is-2">Latest Forum Posts</h2>
                        <div className="columns">
                            {error ? (
                                <div className="notification is-warning">
                                    <p>
                                        There was a problem loading the forum
                                        topics. To view the latest forum topics,{" "}
                                        <a
                                            href="https://forum.codeselfstudy.com/"
                                            target="_blank"
                                        >
                                            visit the forum
                                        </a>
                                        .
                                    </p>
                                    <p>Error message:</p>
                                    <code>{error}</code>
                                </div>
                            ) : null}
                            {topics ? (
                                <Topics topics={topics} />
                            ) : error ? null : (
                                <Spinner
                                    theme="dots2"
                                    size="5em"
                                    style={{ marginTop: "30px" }}
                                />
                            )}
                        </div>
                    </div>
                </div>
            </section>
        </div>
    );
}
