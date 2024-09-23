import { defineConfig } from 'vite';
import isWsl from 'is-wsl';
import react from '@vitejs/plugin-react';

export default defineConfig(() => {
  let watchWSL = isWsl ? {
    watch: {
      usePolling: true,
      interval: 500,
      binaryInterval: 1000,
    }
  } : null;

  return {
    build: {
      outDir: 'build',
    },
    plugins: [react()],
    server: watchWSL,
  };
});
