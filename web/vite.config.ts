import adapter from '@sveltejs/adapter-static';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			adapter: adapter({
				pages: 'build',
				assets: 'build',
				fallback: undefined,
				precompress: true,
				strict: true
			})
		})
	],
	server: {
		proxy: {
			'/api': 'http://localhost:3000',
			'/config': 'http://localhost:3000',
			'/healthz': 'http://localhost:3000'
		}
	}
});
