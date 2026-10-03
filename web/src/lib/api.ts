import { loggedIn } from '#lib/store.svelte.js';

export async function client<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const headers = new Headers(options?.headers);
	headers.set('Content-Type', 'application/json');

	const response = await fetch(`${endpoint}`, {
		...options,
		credentials: 'include',
		headers
	});

	if (!response.ok) {
		if (response.status === 401) loggedIn.current = false;
		const errorBody = await response.text();
		throw new Error(errorBody || `API Error: ${response.status} - ${response.statusText}`);
	}

	// Ensure the user is logged in if the request was successful
	if (!loggedIn.current) loggedIn.current = true;

	const text = await response.text();
	return (text ? JSON.parse(text) : undefined) as T;
}

export type Agent = {
	name: string;
	addr: string;
	connected: boolean;
	since: string;
	updated: string;
	routers: number;
	services: number;
};

export type Collision = {
	kind: string;
	name: string;
	source: string;
	owner: string;
};

export type SharedService = {
	name: string;
	agents: string[];
	servers: string[];
};

export type Snapshot = {
	config: Record<string, any>;
	agents: Agent[] | null;
	collisions: Collision[] | null;
	shared: SharedService[] | null;
};

export const api = {
	login: (secret: string) =>
		client<void>('/api/login', {
			method: 'POST',
			body: JSON.stringify({ secret })
		}),
	logout: async () => {
		await client<void>('/api/logout', { method: 'POST' });
		loggedIn.current = false;
	},
	envs: () => client<string[]>('/api/envs'),

	events(env: string, onSnapshot: (s: Snapshot) => void): EventSource {
		const source = new EventSource(`/api/events?env=${encodeURIComponent(env)}`, {
			withCredentials: true
		});
		source.onmessage = (event) => onSnapshot(JSON.parse(event.data));
		// EventSource hides the status code, a regular request tells us if the session expired.
		source.onerror = () => {
			api.envs().catch(() => {});
		};
		return source;
	}
};
