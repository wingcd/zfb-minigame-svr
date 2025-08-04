import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  base:"./",
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server:{
      host:'0.0.0.0',
      port:3000,
      open:true,
      cors:true,
      proxy: {
        '/admin': {
          target: 'https://env-00jxt0uhcb2h.api-hz.cloudbasefunction.cn',
          changeOrigin: true,
          secure: false
        }
      }
  }
})