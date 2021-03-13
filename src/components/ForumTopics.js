import React, { useEffect } from "react";

import "./ForumTopics.scss";

export default function ForumPosts(props) {
    useEffect(() => {
        async function loadPosts() {
            try {
                const topicLoader = require("@j127/embed-discourse/loader");
                await topicLoader.defineCustomElements(window);
            } catch (err) {
                console.error(err);
            }
        }
        loadPosts();
    }, []);

    return (
        <div className="forum-posts">
            <section className="section">
                <div className="container content">
                    <div className="box">
                        <h2 className="title is-2">Latest Forum Posts</h2>
                        <div className="columns">
                            <div className="column">
                                <discourse-embed-topics forum-base-url="https://forum.codeselfstudy.com"></discourse-embed-topics>
                            </div>
                            <div className="column">
                                <discourse-embed-topics
                                    forum-base-url="https://forum.codeselfstudy.com"
                                    offset="7"
                                ></discourse-embed-topics>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    );
}
