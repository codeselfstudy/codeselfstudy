import React from "react";

const styles = {
    fontSize: "5rem",
};

/**
 * A spinning loader, entirely in text.
 *
 * Options: "loading", "loading dots", or "loading dots2"
 */
export default function Spinner({ kind = "loading dots" }) {
    return <span style={styles} className={kind} title="loading"></span>;
}
