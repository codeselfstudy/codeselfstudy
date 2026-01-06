import { useEffect, useState } from "react";

/**
 * Displays the current year.
 *
 * This component is implemented with a client-side effect to ensure that any
 * server-rendered or statically generated content is updated with the current
 * date upon hydration.
 */
export default function CurrentYear() {
  const [year, setYear] = useState(new Date().getFullYear());

  useEffect(() => {
    const y = new Date().getFullYear();
    setYear(y);
  }, []);

  return <span>{year}</span>;
}
