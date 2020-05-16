import React from "react";

const StepCheckbox = ({ step, label }) => {
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
            </div>
        </div>
    );
};

export default StepCheckbox;
