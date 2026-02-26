import "./ssr.polyfill.ts";

import { createInertiaApp } from "@inertiajs/react";
import { renderToString } from "react-dom/server.edge";

const pages = import.meta.glob("./pages/**/*.tsx", { eager: true });

export async function renderPage(pageObject: any) {
  const res = await createInertiaApp({
    page: pageObject,
    render: renderToString,
    resolve: (name) => {
      return pages[`./pages/${name}.tsx`];
    },
    setup: ({ App, props }) => <App {...props} />,
  });

  return res;
}
