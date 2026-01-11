import { Plugin } from "vite";

type Config = {
  entryPoint: string;
  ssrEntryPoint?: string;
};

export function goSsrPlugin({ entryPoint, ssrEntryPoint }: Config): Plugin[] {
  return [
    {
      name: "go-ssr-dev-middleware",
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
      configureServer(server) {
        if (!ssrEntryPoint) {
          return;
        }

        server.middlewares.use("/render", async (req, res) => {
          try {
            const body = (await readBody(req)) as any;

            const pageObject = JSON.parse(body);

            const { renderPage } = await server.ssrLoadModule(ssrEntryPoint);

            if (!renderPage) {
              throw new Error(
                `Export 'renderPage' not found in ${ssrEntryPoint}`,
              );
            }

            const renderResult = await renderPage(pageObject);

            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify(renderResult));
          } catch (e: any) {
            // Fix the stack trace so it points to your actual source files
            server.ssrFixStacktrace(e);
            console.error("SSR Middleware Error:", e);

            res.statusCode = 500;
            res.end(
              JSON.stringify({
                error: e.message,
                stack: e.stack,
              }),
            );
          }
        });
      },
    },
  ];
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    let data = "";
    req.on("data", (chunk) => {
      data += chunk;
    });
    req.on("end", () => {
      resolve(data);
    });
    req.on("error", (err) => {
      reject(err);
    });
  });
}
