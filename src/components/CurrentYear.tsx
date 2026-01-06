import { useEffect, useState } from "react";

export default function CurrentYear() {
  const [year, setYear] = useState(new Date().getFullYear());

  useEffect(() => {
    const y = new Date().getFullYear();
    setYear(y);
  }, []);

  return <span>{year}</span>;
}
