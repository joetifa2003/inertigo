import { Link } from "@inertiajs/react";
import { useState } from "react";

export default function Index(props) {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>Home</h1>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
      <br />
      <Link href="/about">Go To About</Link>
    </div>
  );
}
