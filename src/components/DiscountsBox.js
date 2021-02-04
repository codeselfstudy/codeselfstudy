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
        href: "/discount/digital-ocean/",
    },
    {
        id: 2,
        name: "AlgoExpert",
        src: algoExpertLogo,
        href: "/discount/algoexpert/",
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

export default function DiscountsBox() {
    const [sponsors, setDiscounts] = useState([]);

    // Randomize the sponsors in the box.
    useEffect(() => {
        setDiscounts(shuffle(sponsorsData));
    }, []);

    return (
        <div className="box">
            <h2 className="title is-2">Discounts</h2>
            <p>
                These companies provide services, discounts, or other assistance
                to our group members.
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
            <p>
                If you would like to help offer discounts to our members or
                sponsor the group in some way, please{" "}
                <Link to="/contact/">contact us</Link>.
            </p>
        </div>
    );
}
