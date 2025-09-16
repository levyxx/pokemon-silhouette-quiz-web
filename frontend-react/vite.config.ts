import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // 既に 5173 使用中なら自動で別ポートへ (固定したいなら strictPort:true)
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 必要ならパスを書き換える例:
        // rewrite: path => path.replace(/^\/api/, '/api')
      }
    }
  }
});