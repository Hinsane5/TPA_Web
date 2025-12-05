import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

export default defineConfig({
  plugins: [vue(), vueDevTools()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },

  server: {
    proxy: {
      "/api/chats": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },

      "/api/v1": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },

      // "/api/stories": {
      //   target: "http://localhost:8081",
      //   changeOrigin: true,
      // },

      "/api": {
        target: "http://localhost:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
