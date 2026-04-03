import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // 代理 TWSE 即時報價 API (加權指數、個股等)
      '/api/twse': {
        target: 'https://mis.twse.com.tw',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/twse/, '/stock/api'),
        headers: {
          'Referer': 'https://mis.twse.com.tw/',
          'Origin': 'https://mis.twse.com.tw',
        },
      },
      // 代理 TWSE OpenAPI (收盤統計資料)
      '/api/twse-open': {
        target: 'https://openapi.twse.com.tw',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/twse-open/, ''),
      },
      // 代理 TAIFEX 期交所資料
      '/api/taifex': {
        target: 'https://mis.taifex.com.tw',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/taifex/, ''),
        headers: {
          'Referer': 'https://mis.taifex.com.tw/',
          'Origin': 'https://mis.taifex.com.tw',
        },
      },
    },
  },
})
