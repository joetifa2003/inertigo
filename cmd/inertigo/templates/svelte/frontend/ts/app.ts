import { createInertiaApp } from "@inertiajs/svelte";
import { hydrate } from "svelte";
import "../styles/app.css";

const pages = import.meta.glob("./pages/**/*.svelte");

createInertiaApp({
    resolve: (name) => {
        return pages[`./pages/${name}.svelte`]();
    },
    setup({ el, App, props }) {
        hydrate(App, { target: el, props });
    },
});
