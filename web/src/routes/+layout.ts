import { api } from '#lib/api.js';
import { loggedIn } from '#lib/store.svelte.js';

export const ssr = false;
export const prerender = true;

// Skips the login screen when the session is still valid or no token is set.
export async function load() {
	if (!loggedIn.current) await api.envs().catch(() => {});
}
