import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api/v1/auth": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },
      "/api/v1/users": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },
      "/api/v1/wallet": {
        target: "http://localhost:8082",
        changeOrigin: true,
      },
      "/api/v1/orders": {
        target: "http://localhost:8083",
        changeOrigin: true,
      },
    },
  },
});
