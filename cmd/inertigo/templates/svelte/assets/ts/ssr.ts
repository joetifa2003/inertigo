import "./ssr.polyfill.ts";

import { createInertiaApp } from "@inertiajs/svelte";
import { render } from "svelte/server";

const pages = import.meta.glob("./pages/**/*.svelte", { eager: true });

export async function renderPage(page: any) {
  const res = await createInertiaApp({
    page,
    resolve: (name) => {
      return pages[`./pages/${name}.svelte`];
    },
    setup({ App, props }) {
      return render(App, { props });
    },
  });

  return res;
}
