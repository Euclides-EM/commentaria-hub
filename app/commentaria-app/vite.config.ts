import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

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
