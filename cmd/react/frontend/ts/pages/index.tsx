import { Deferred, Link } from "@inertiajs/react";
import { PageIndex } from "typegen";

export default function Index({ message, reviews }: PageIndex) {
  return (
    <div>
      <div>{message}</div>
      <Deferred data="reviews" fallback={<div>Loading reviews...</div>}>
        {reviews?.map((review) => (
          <div key={review}>{review}</div>
        ))}
      </Deferred>
      <Link href="/register">Register</Link>
    </div>
  );
}
