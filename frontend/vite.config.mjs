import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// DEV stays fixed — no templating
const DEV_ALLOWED = ['frontend', 'localhost', '127.0.0.1']

// PROD is injected at build-time by scripts/prod/env_writer_prod.sh
// It becomes a concrete JS array literal like: ["brierfoxforecast.com", "www.brierfoxforecast.com"]
const PROD_ALLOWED = ["social.ntoufoudis.com", "www.social.ntoufoudis.com"]

export default defineConfig(({ mode }) => {
  const isProd = mode === 'production'
  const allowed = isProd ? PROD_ALLOWED.filter(Boolean) : DEV_ALLOWED

  return {
    server: {
      host: '0.0.0.0',
      allowedHosts: allowed,
    },
    preview: { allowedHosts: allowed },
    build: {
      outDir: 'build',
      commonjsOptions: { transformMixedEsModules: true },
    },
    plugins: [react()],
    css: { target: 'async' },
  }
})
