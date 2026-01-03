import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { goSsrPlugin } from "./plugin";

export default defineConfig(({ isSsrBuild }) => ({
  plugins: [
    react(),
    goSsrPlugin({
      entryPoint: "./ts/app.tsx",
      ssrEntryPoint: "./ts/ssr.tsx",
    }),
  ],
}));
