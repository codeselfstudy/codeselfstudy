import React from "react";
import { Link } from "gatsby";
import digitaloceanLogo from "../images/sponsors/digitalocean-sq.png";
import fullstackRustLogo from "../images/sponsors/fullstack-rust.png";

const sponsors = [
    {
        id: 1,
        name: "Digital Ocean",
        src: digitaloceanLogo,
        href: "/sponsor/digital-ocean/",
    },
];

function Sponsor({ name, src, href }) {
    return (
        <div style={{ margin: "0 11px" }}>
            {/^http.*/.test(href) ? (
                <a
                    href={href}
                    title={name}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <img src={src} alt={`${name} logo`} />
                </a>
            ) : (
                <Link to={href} title={name}>
                    <img src={src} alt={`${name} logo`} />
                </Link>
            )}
        </div>
    );
}

export default function SponsorsBox() {
    return (
        <section className="section">
            <div className="container content">
                <div className="box">
                    <h2 className="title is-2">Sponsors</h2>
                    <p>
                        These companies provide assistance to our group. To
                        support us and add your logo, please{" "}
                        <Link to="/contact/">contact us</Link>.
                    </p>
                    <div className="columns" style={{ margin: "27px 17px" }}>
                        {sponsors.map((s) => (
                            <Sponsor
                                className="column"
                                key={s.id}
                                name={s.name}
                                src={s.src}
                                href={s.href}
                            />
                        ))}
                    </div>
                </div>
            </div>
        </section>
    );
}
