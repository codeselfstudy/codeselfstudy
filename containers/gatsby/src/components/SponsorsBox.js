import React, { useEffect, useState } from "react";
import { Link } from "gatsby";
import { shuffle } from "lodash";
import digitalOceanLogo from "../images/sponsors/digital-ocean.png";
import algoExpertLogo from "../content/files/algoexpert-logo.png";

const sponsorsData = [
    {
        id: 1,
        name: "Digital Ocean",
        src: digitalOceanLogo,
        href: "/sponsor/digital-ocean/",
    },
    {
        id: 2,
        name: "AlgoExpert",
        src: algoExpertLogo,
        href: "/sponsor/algoexpert/",
    },
];

const sponsorStyles = {
    margin: "17px",
};

function Sponsor({ name, src, href }) {
    return (
        <div className="sponsor" style={sponsorStyles}>
            <Link to={href} title={name}>
                <img src={src} alt={`${name} logo`} />
            </Link>
        </div>
    );
}

export default function SponsorsBox() {
    const [sponsors, setSponsors] = useState([]);

    // Randomize the sponsors in the box.
    useEffect(() => {
        setSponsors(shuffle(sponsorsData));
    }, []);

    return (
        <div className="box">
            <h2 className="title is-2">Sponsors</h2>
            <p>
                These companies provide services, discounts, or other assistance
                to our group. To support us and add your logo, please{" "}
                <Link to="/contact/">contact us</Link>.
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
