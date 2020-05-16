import React from "react";

const StepCheckbox = ({ step, label, okToSkip = false }) => {
    return (
        <div className="field">
            <div className="control">
                <div>
                    <label className="radio">
                        <input type="radio" name="answer" />{" "}
                        <strong>{label}</strong>
                    </label>
                </div>
                <div>
                    <label className="radio">
                        <input type="radio" name="answer" /> This item isn't
                        relevant to my problem.
                    </label>
                </div>
                {okToSkip ? (
                    <div>
                        <label className="radio">
                            <input type="radio" name="answer" /> I don't know
                            how to do this.
                        </label>
                    </div>
                ) : null}
            </div>
            <button className="button is-link">Next</button>
        </div>
    );
};

export default StepCheckbox;
