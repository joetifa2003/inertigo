import { useState } from "react";

export default function Index({ message }: { message: string }) {
  const [count, setCount] = useState(0);

  return (
    <div style={{ padding: "20px", fontFamily: "sans-serif" }}>
      <h1>Inertigo Example</h1>
      <p>Message from Go: <strong>{message}</strong></p>

      <div style={{ marginTop: "20px" }}>
        <p>Counter: {count}</p>
        <button onClick={() => setCount(count + 1)}>Increment</button>
      </div>
    </div>
  );
}
