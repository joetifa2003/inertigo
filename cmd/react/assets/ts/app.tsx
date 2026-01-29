import { createInertiaApp } from "@inertiajs/react";
import { hydrateRoot } from "react-dom/client";
import "../styles/app.css";

const pages = import.meta.glob("./pages/**/*.tsx", { eager: true });

createInertiaApp({
  resolve: (name) => {
    return pages[`./pages/${name}.tsx`]();
  },
  setup({ el, App, props }) {
    hydrateRoot(el, <App {...props} />);
  },
});
