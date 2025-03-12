import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import generouted from "@generouted/react-router/plugin";

// https://vite.dev/config/
export default defineConfig({
    plugins: [react(), tailwindcss(), generouted()],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        watch: {
          usePolling: true
        },
        proxy: {
          '/api': {
            target: 'http://localhost:3000',
            changeOrigin: true,
            secure: false,
            rewrite: (path) => path.replace(/^\/api/, '')
          },
        }
      }
});
