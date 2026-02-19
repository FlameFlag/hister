import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    output: {
      bundleStrategy: 'single'
    },
    adapter: adapter({
      pages: 'dist',
      assets: 'dist',
      fallback: 'index.html',
      precompress: true,
    }),
    alias: {
      $lib: './src/lib',
      $components: './src/lib/components',
    },
    prerender: {
      entries: [],
    },
  },
  typescript: {
    config(config) {
      config.exclude = [...(config.exclude || []), 'node_modules', 'dist', '.svelte-kit'];
      return config;
    },
  },
};

export default config;
