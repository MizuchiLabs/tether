import { api, type Snapshot } from '#lib/api.js';
import { env } from '#lib/store.svelte.js';

class Live {
	envs = $state.raw<string[] | null>(null);
	snapshot = $state.raw<Snapshot | null>(null);
	offline = $state(false);

	async refreshEnvs() {
		try {
			const names = await api.envs();
			this.envs = names;
			if (names.length > 0 && !names.includes(env.current)) {
				env.current = names.includes('default') ? 'default' : names[0];
			}
		} catch {
			// A 401 already logs out, anything else gets retried on the next poll.
		}
	}

	/** Streams snapshots for one environment. Returns a cleanup function. */
	connect(name: string) {
		let source: EventSource;
		let retry: ReturnType<typeof setTimeout> | undefined;

		const open = () => {
			source = new EventSource(`/api/events?env=${encodeURIComponent(name)}`, {
				withCredentials: true
			});
			source.onmessage = (event) => {
				this.snapshot = JSON.parse(event.data);
				this.offline = false;
			};
			source.onerror = () => {
				this.offline = true;
				// EventSource hides the status code, a regular request tells us if the session expired.
				api.envs().catch(() => {});
				// It retries dropped connections itself but gives up on error responses, like a 502 from a proxy.
				if (source.readyState === EventSource.CLOSED) retry = setTimeout(open, 3000);
			};
		};

		this.snapshot = null;
		open();

		return () => {
			clearTimeout(retry);
			source.close();
			this.offline = false;
		};
	}
}

export const live = new Live();
