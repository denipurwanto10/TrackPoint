import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // src/ berisi arsip .zip yang terkunci proses lain; watcher crash (EBUSY) tanpa ini
    watch: { ignored: ["**/*.zip"] },
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
