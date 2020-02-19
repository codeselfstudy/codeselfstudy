import React from "react";
import { Link } from "gatsby";
import digitalOceanLogo from "../images/sponsors/digital-ocean.png";

const sponsors = [
    {
        id: 1,
        name: "Digital Ocean",
        src: digitalOceanLogo,
        href: "/sponsor/digital-ocean/",
    },
];

function Sponsor({ name, src, href }) {
    return (
        <Link to={href} title={name}>
            <img src={src} alt={`${name} logo`} />
        </Link>
    );
}

export default function SponsorsBox() {
    return (
        <div className="box">
            <h2 className="title is-2">Sponsors</h2>
            <p>
                These companies provide assistance to our group. To support us
                and add your logo, please <Link to="/contact/">contact us</Link>
                .
            </p>
            <div className="columns" style={{ margin: "27px 17px" }}>
                {sponsors.map(s => (
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
    );
}
