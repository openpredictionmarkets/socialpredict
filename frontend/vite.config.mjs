import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(() => {
  return {
    build: {
      outDir: 'build',
      commonjsOptions: { transformMixedEsModules: true },
    },
    plugins: [react()],
    css: {
      target: 'async', // or 'defer' for older browsers
    },
  };
});
