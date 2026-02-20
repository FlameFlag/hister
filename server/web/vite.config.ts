import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [sveltekit(), tailwindcss()],
  resolve: {
    alias: {
      $lib: './src/lib',
      $components: './src/lib/components',
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: () => 'bundle.js',
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  optimizeDeps: {
    rolldownOptions: {
      exclude: [],
    },
  },
  kit: {
    output: {
      inlineStyleThreshold: Infinity,
    },
  },
});
