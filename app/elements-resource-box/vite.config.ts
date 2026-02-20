import type { ViteDevServer } from "vite";
import { defineConfig, loadEnv, Plugin } from "vite";
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import path from "node:path";

const normalizeBasePath = (value?: string): string => {
  const raw = (value ?? "").trim();
  if (!raw || raw === "/") {
    return "/";
  }
  const withLeadingSlash = raw.startsWith("/") ? raw : `/${raw}`;
  return withLeadingSlash.endsWith("/")
    ? withLeadingSlash
    : `${withLeadingSlash}/`;
};

function basePathRedirectPlugin(base: string): Plugin {
  return {
    name: "base-path-redirect",
    configureServer(server: ViteDevServer) {
      if (base === "/") {
        return;
      }
      const baseNoTrailing = base.slice(0, -1);
      server.middlewares.use((req, res, next) => {
        if (req.method === "GET" && req.url === baseNoTrailing) {
          res.statusCode = 302;
          res.setHeader("Location", base);
          res.end();
          return;
        }
        next();
      });
    },
  };
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const base = normalizeBasePath(env.VITE_BASE_PATH);

  return {
    base,
    resolve: {
      alias: {
        "@hub-api": path.resolve(__dirname, "../hub-api/hub-api/index.ts"),
      },
    },
    plugins: [react(), svgr(), basePathRedirectPlugin(base)],
    server: {
      port: 5190,
      allowedHosts: ["euclides.huma-num.fr"],
    },
  };
});
