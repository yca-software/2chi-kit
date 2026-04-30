import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    preserveSymlinks: false,
    // Workspace design-system; ensure a single React instance
    dedupe: ["react", "react-dom"],
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
    extensions: [".mjs", ".js", ".mts", ".ts", ".jsx", ".tsx", ".json"],
  },
  server: {
    host: "0.0.0.0",
    allowedHosts: ["localhost", "127.0.0.1"],
    port: 3000,
    open: true,
    fs: {
      strict: false,
    },
  },
  optimizeDeps: {
    include: [
      "@tanstack/react-query",
      "react",
      "react-dom",
      "react-router",
      "zustand",
      "use-sync-external-store/shim",
      "use-sync-external-store/shim/with-selector",
    ],
    // Keep linked workspace package out of pre-bundling so style/class
    // updates in `packages/design-system` are reflected during local dev.
    exclude: ["@yca-software/design-system"],
  },
  build: {
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: undefined,
      },
    },
  },
  esbuild: {
    sourcemap: false,
  },
});
