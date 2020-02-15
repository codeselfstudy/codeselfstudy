import React, { useState } from 'react';

import Layout from '../../components/Layout';
import SEO from '../../components/SEO';

import { sansSerif } from '../../data/unicode';

export default function Unicode() {
    const [text, setText] = useState('');
    const [output, setOutput] = useState('');

    function handleOnKeyUp(e) {
        const chars = e.target.value;
        const translated = [...chars]
            .map(char => {
                console.log('char', char);
                return sansSerif[char] || char;
            })
            .join('');
        setText(chars);
        setOutput(translated);
    }
    return (
        <Layout>
            <SEO title="Bold Unicode Text Tool" />

            <section className="section">
                <div className="container content">
                    <h1 className="title is-1">Bold Unicode Text Tool</h1>
                    <p>Type your text below to get the unicode characters.</p>

                    <div class="field">
                        <div class="control">
                            <input
                                class="input"
                                type="text"
                                onKeyUp={handleOnKeyUp}
                                placeholder="Enter some text"
                            />
                        </div>
                    </div>
                    {text ? (
                        <div className="box">
                            <small
                                class="is-size-7"
                                style={{ fontFamily: 'monospace' }}
                            >
                                OUTPUT:
                            </small>
                            <br />
                            {output}
                        </div>
                    ) : (
                        <></>
                    )}
                </div>
            </section>
        </Layout>
    );
}
