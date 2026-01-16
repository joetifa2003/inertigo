import type { Plugin, ViteDevServer } from "vite";
import type { IncomingMessage } from "http";

export type InertiaPageObject = {
    component: string;
    props: Record<string, unknown>;
    url: string;
    version: string;
    encryptHistory?: boolean;
    clearHistory?: boolean;
    [key: string]: unknown;
};

export type InertigoConfig = {
    /**
     * The entry point for the client-side application.
     * This is typically your main React/Vue/Svelte entry file.
     */
    entryPoint: string;

    /**
     * The entry point for server-side rendering (optional).
     * If provided, a `/render` endpoint will be available during development
     * for SSR rendering via the Go backend.
     */
    ssrEntryPoint?: string;
};

/**
 * Vite plugin for Inertigo - The Inertia.js adapter for Go.
 *
 * This plugin configures Vite for use with Inertigo, providing:
 * - CORS support for development server
 * - Development hot reloading ssr
 * - Build manifest generation
 * - Development SSR middleware (when ssrEntryPoint is provided)
 *
 * @example
 * ```ts
 * // vite.config.ts
 * import { defineConfig } from 'vite';
 * import react from '@vitejs/plugin-react';
 * import inertigo from 'vite-plugin-inertigo';
 *
 * export default defineConfig({
 *   plugins: [
 *     react(),
 *     inertigo({
 *       entryPoint: './src/main.tsx',
 *       ssrEntryPoint: './src/ssr.tsx', // optional
 *     }),
 *   ],
 * });
 * ```
 */
export default function inertigo({ entryPoint, ssrEntryPoint }: InertigoConfig): Plugin[] {
    return [
        {
            name: "inertigo",
            config(_, { isSsrBuild }) {
                return {
                    server: {
                        cors: true,
                    },
                    ssr: {
                        target: "webworker",
                        noExternal:
                            process.env.NODE_ENV === "production" ? true : undefined,
                    },
                    build: {
                        target: isSsrBuild ? "es2023" : undefined,
                        manifest: true,
                        rollupOptions: {
                            input: {
                                app: entryPoint,
                            },
                        },
                        commonjsOptions: {
                            transformMixedEsModules: true,
                        },
                    },
                };
            },
            configureServer(server: ViteDevServer) {
                if (!ssrEntryPoint) {
                    return;
                }

                server.middlewares.use("/render", async (req, res) => {
                    try {
                        const body = await readBody(req);
                        const pageObject: InertiaPageObject = JSON.parse(body);

                        const { renderPage } = await server.ssrLoadModule(ssrEntryPoint);

                        if (!renderPage) {
                            throw new Error(
                                `Export 'renderPage' not found in ${ssrEntryPoint}`,
                            );
                        }

                        const renderResult = await renderPage(pageObject);

                        res.setHeader("Content-Type", "application/json");
                        res.end(JSON.stringify(renderResult));
                    } catch (e: unknown) {
                        const error = e as Error;
                        // Fix the stack trace so it points to your actual source files
                        server.ssrFixStacktrace(error);
                        console.error("SSR Middleware Error:", error);

                        res.statusCode = 500;
                        res.end(
                            JSON.stringify({
                                error: error.message,
                                stack: error.stack,
                            }),
                        );
                    }
                });
            },
        },
    ];
}

function readBody(req: IncomingMessage): Promise<string> {
    return new Promise((resolve, reject) => {
        let data = "";
        req.on("data", (chunk: Buffer | string) => {
            data += chunk;
        });
        req.on("end", () => {
            resolve(data);
        });
        req.on("error", (err: Error) => {
            reject(err);
        });
    });
}


