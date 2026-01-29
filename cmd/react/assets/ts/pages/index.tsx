import { Deferred, Link, router, usePage } from "@inertiajs/react";
import { useEffect } from "react";

export default function Index({ message, reviews, date }) {
  const page = usePage();

  return (
    <div>
      <div>{message}</div>
      <Deferred data="reviews" fallback={<div>Loading reviews...</div>}>
        {reviews?.map((review) => (
          <div key={review}>{review}</div>
        ))}
      </Deferred>
      <div>Date: {date}</div>
      <button
        onClick={() => {
          router.reload({
            only: ["date"],
          });
        }}
      >
        Reload Date
      </button>
      <Link href="/register">Register</Link>
    </div>
  );
}
