import "./ssr.polyfill.ts";

import { createInertiaApp } from "@inertiajs/vue3";
import createServer from "@inertiajs/vue3/server";
import { renderToString } from "@vue/server-renderer";
import { createSSRApp, h } from "vue";

const pages = import.meta.glob("./pages/**/*.vue", { eager: true });

export async function renderPage(pageObject: any) {
  const res = await createInertiaApp({
    page: pageObject,
    render: renderToString,
    resolve: (name) => {
      return pages[`./pages/${name}.vue`];
    },
    setup({ App, props, plugin }) {
      return createSSRApp({
        render: () => h(App, props),
      }).use(plugin);
    },
  });

  return res;
}
