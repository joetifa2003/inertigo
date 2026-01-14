import { defineConfig } from "vite";
import { goSsrPlugin } from "./plugin";
import vue from "@vitejs/plugin-vue";

export default defineConfig(({ isSsrBuild }) => ({
  plugins: [
    vue(),
    goSsrPlugin({
      entryPoint: "./ts/app.ts",
      ssrEntryPoint: "./ts/ssr.ts",
    }),
  ],
  server: {
    port: 5174,
  }
}));
