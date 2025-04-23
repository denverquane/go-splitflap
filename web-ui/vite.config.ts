import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import generouted from "@generouted/react-router/plugin";

const env = loadEnv(process.env.NODE_ENV as string, process.cwd(), 'VITE_');

const API_URL = env.VITE_BACKEND_API_URL || "http://localhost:3000"

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
            target: API_URL,
            changeOrigin: true,
            secure: false,
            rewrite: (path) => path.replace(/^\/api/, '')
          },
        }
      }
});
