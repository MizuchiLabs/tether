import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/config': 'http://localhost:3000',
			'/envs': 'http://localhost:3000',
			'/healthz': 'http://localhost:3000',
		}
	}
});
