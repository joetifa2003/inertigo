import { Deferred, Link, useRemember, usePage } from "@inertiajs/react";

export default function Index(props) {
  const [count, setCount] = useRemember(0);
  const {
    props: { auth, appName, year },
    flash,
  } = usePage() as any;

  return (
    <div>
      <h1>Home</h1>

      {/* Show flash messages */}
      {flash?.success && (
        <div
          style={{
            padding: "10px",
            background: "#4caf50",
            color: "white",
            marginBottom: "10px",
            borderRadius: "4px",
          }}
        >
          ✓ {flash.success}
        </div>
      )}

      {/* Show shared props */}
      {auth.user && (
        <div
          style={{
            padding: "10px",
            background: "#e3f2fd",
            marginBottom: "10px",
            borderRadius: "4px",
          }}
        >
          Logged in as: {auth.user.name} ({auth.user.email})
        </div>
      )}

      {appName && (
        <p style={{ color: "#666", fontSize: "14px" }}>
          {appName} - {year}
        </p>
      )}

      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
      <br />
      <Link href="/about">Go To About</Link>
      <Deferred data="lazyMessage" fallback={<div>Loading...</div>}>
        <p>Lazy Message: {props.lazyMessage}</p>
      </Deferred>
    </div>
  );
}
