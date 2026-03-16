import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import VueRouter from 'vue-router/vite'
import VueDevtools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    VueRouter(),
    VueDevtools(),
    tailwindcss(),
    vue(),
  ],
  server: {
    proxy: {
      // Proxy all /api requests to your backend
      '/api': {
        target: 'http://localhost:8080', // your backend URL
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''), // optional: remove /api prefix
      },
    },
  },
});