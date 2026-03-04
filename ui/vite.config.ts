import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { viteSingleFile } from 'vite-plugin-singlefile'

export default defineConfig({
  plugins: [
    react(),
    viteSingleFile(),
  ],
  build: {
    // Disable source maps for smaller output
    sourcemap: false,
    // Inline all assets
    assetsInlineLimit: Infinity,
    // Single chunk output
    rollupOptions: {
      output: {
        manualChunks: undefined,
      },
    },
    // Minimize aggressively
    minify: 'esbuild',
    // Target modern browsers only (reduces polyfill size)
    target: 'esnext',
  },
})
