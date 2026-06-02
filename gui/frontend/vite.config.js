import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
  ],
  test: {
    environment: 'jsdom',
    css: false,
    include: ['src/**/*.test.js'],
  },
})