import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(() => {
  return {
    server: {
      allowedHosts: [
        'frontend', // we need this to be able to access the app on localhost
	'localhost',
	'brierfoxforecast.com', // staging instance
	'openprediction.xyz' // prod instance

	// add your own domain here!

      ]
    },
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
