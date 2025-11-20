import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },

  server: {
    proxy: {
      // 1. SPECIFIC rule for /api/v1 (for Posts, etc.)
      // This matches first and DOES NOT rewrite.
      "/api/chats": {
        target: "http://localhost:8081",
        changeOrigin: true,
        // DO NOT use rewrite here. The Gateway expects /api/chats
      },

      // 2. Posts Service
      "/api/v1": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },

      // --- GENERAL RULE LAST (Rewrite enabled) ---

      // 3. Auth Service (Catch-all for other /api requests)
      "/api": {
        target: "http://localhost:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
