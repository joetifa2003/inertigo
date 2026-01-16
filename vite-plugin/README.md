# inertigo-vite

Vite plugin for [Inertigo](https://github.com/joetifa2003/inertigo) - The Inertia.js adapter for Go.

## Installation

```bash
npm install inertigo-vite --save-dev
# or
pnpm add inertigo-vite -D
# or
yarn add inertigo-vite --dev
```

## Usage

```ts
// vite.config.ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import inertigo from "inertigo-vite";

export default defineConfig({
  plugins: [
    react(),
    inertigo({
      entryPoint: "./src/main.tsx",
      ssrEntryPoint: "./src/ssr.tsx", // optional, for SSR support
    }),
  ],
});
```

## Configuration

### `entryPoint` (required)

The entry point for your client-side application. This is typically your main React/Vue/Svelte entry file.

### `ssrEntryPoint` (optional)

The entry point for server-side rendering. If provided, a `/render` endpoint will be available during development for SSR rendering via the Go backend.

## Features

This plugin configures Vite for use with Inertigo, providing:

- Basic starting point for inertia bundling.
- Development SSR middleware (when ssrEntryPoint is provided)

## SSR Entry Point

Your SSR entry point should export a `renderPage` function:

```tsx
// src/ssr.tsx
import { createInertiaApp } from "@inertiajs/react";
import ReactDOMServer from "react-dom/server";

export async function renderPage(page) {
  return createInertiaApp({
    page,
    render: ReactDOMServer.renderToString,
    resolve: (name) => {
      const pages = import.meta.glob("./pages/**/*.tsx", { eager: true });
      return pages[`./pages/${name}.tsx`];
    },
    setup: ({ App, props }) => <App {...props} />,
  });
}
```