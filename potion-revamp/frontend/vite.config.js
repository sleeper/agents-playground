import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/pages': 'http://localhost:8080',
      '/databases': 'http://localhost:8080'
    }
  },
  test: {
    environment: 'jsdom',
    setupFiles: './vitest.setup.js'
  }
});
