import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(() => {
  return {
    build: {
      outDir: 'build',
      commonjsOptions: { transformMixedEsModules: true }
    },
    plugins: [ react(), ],
    server: {   //dev mode
        host: "192.168.1.6",    //frontend server addr and port. should be added to backend AllowedOrigins
        port: 8089,               
        proxy: {
                    '/api/v0': {
                    target: 'http://localhost:8080',//backend server addr and port
                    changeOrigin: true,
                    ws:false,
                    rewrite: (pathStr)=>pathStr.replace('/api/v0', '/v0'),// url rewrite
                    timeout:5000,
                },
        },
    }
  };
});
