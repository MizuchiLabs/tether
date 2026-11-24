import { api } from '#lib/api.js';
import { loggedIn } from '#lib/store.svelte.js';

export const ssr = false;
export const prerender = true;

// Keeps the last known state when Tether can't be reached.
export async function load() {
	const session = await api.session().catch(() => null);
	if (session) loggedIn.current = session.signed_in;
	return { mode: session?.mode ?? 'password' };
}
