import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const rawBase = env.VITE_BASE_PATH?.trim()

  let base = '/'
  if (rawBase) {
    base = rawBase.startsWith('/') ? rawBase : `/${rawBase}`
    if (!base.endsWith('/')) base = `${base}/`
  }

  return {
    base,
    resolve: {
      alias: {
        '@hub-api': path.resolve(__dirname, '../hub-api/hub-api/index.ts'),
      },
    },
    server: {
      port: 5180,
    },
    plugins: [
      tailwindcss(),
      react({
        babel: {
          plugins: [['babel-plugin-react-compiler']],
        },
      }),
    ],
  }
})
