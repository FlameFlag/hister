import adapter from '@sveltejs/adapter-static';
import { mdsvex } from 'mdsvex';

/** @type {import('@sveltejs/kit').Config} */
export default {
  extensions: ['.svelte', '.md', '.svx'],
  preprocess: [mdsvex({ extensions: ['.md', '.svx'] })],
  kit: {
    adapter: adapter({ pages: 'build', assets: 'build', fallback: undefined }),
    prerender: { handleHttpError: 'warn', handleMissingId: 'ignore' }
  }
};
