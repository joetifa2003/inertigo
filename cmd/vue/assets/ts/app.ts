import { createApp, h } from "vue";
import { createInertiaApp } from "@inertiajs/vue3";
import "../styles/app.css";

const pages = import.meta.glob("./pages/**/*.vue");

createInertiaApp({
  resolve: (name) => {
    return pages[`./pages/${name}.vue`]();
  },
  setup({ el, App, props, plugin }) {
    createApp({ render: () => h(App, props) })
      .use(plugin)
      .mount(el);
  },
});
