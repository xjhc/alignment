import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import dotenv from 'dotenv'

dotenv.config();

export default defineConfig(({ mode }) => {
  const apiUrl = process.env.VITE_API_URL || 'http://localhost:8080';

  return {
    plugins: [react()],
    server: {
      port: 5173,
      proxy: {
        '/api': apiUrl,
        '/ws': {
          target: apiUrl.replace(/^http/, 'ws'),
          ws: true,
        },
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: true,
    },
  };
});
