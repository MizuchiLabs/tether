import { loggedIn } from '#lib/store.svelte.js';

export class ApiError extends Error {
	constructor(
		readonly status: number,
		message: string
	) {
		super(message);
	}
}

export async function client<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const headers = new Headers(options?.headers);
	headers.set('Content-Type', 'application/json');

	const response = await fetch(endpoint, {
		...options,
		credentials: 'include',
		headers
	});

	if (!response.ok) {
		if (response.status === 401) loggedIn.current = false;
		const body = (await response.text()).trim();
		throw new ApiError(response.status, body || response.statusText || `Error ${response.status}`);
	}

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

// Named entries per section, like http.routers.<name>.
export type Sections = Record<string, Record<string, unknown>>;

export type TraefikConfig = {
	http?: Sections;
	tcp?: Sections;
	udp?: Sections;
	tls?: Record<string, unknown>;
};

export type Snapshot = {
	config: TraefikConfig;
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
	envs: async () => (await client<string[] | null>('/api/envs')) ?? []
};
