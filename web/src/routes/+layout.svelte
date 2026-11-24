<script lang="ts">
	import Login from '#lib/components/Login.svelte';
	import AppFooter from '#lib/components/nav/AppFooter.svelte';
	import AppHeader from '#lib/components/nav/AppHeader.svelte';
	import * as Alert from '#lib/components/ui/alert/index.js';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import { loggedIn } from '#lib/store.svelte.js';
	import { TriangleAlertIcon } from '@lucide/svelte';
	import { ModeWatcher } from 'mode-watcher';
	import './layout.css';

	const { children } = $props();
</script>

<!-- Same values as --background in layout.css. -->
<ModeWatcher themeColors={{ dark: '#181825', light: '#eff1f5' }} />
<Toaster richColors />

{#if loggedIn.current}
	<div class="flex min-h-svh flex-col px-4">
		<AppHeader />
		<Alert.Root class="mx-auto mb-6 max-w-5xl">
			<TriangleAlertIcon />
			<Alert.Title>Sign in changes in v0.8.0</Alert.Title>
			<Alert.Description>
				<p>
					<code>TETHER_TOKEN</code> goes away. The web UI signs in with <code>TETHER_PASSWORD</code>
					(12+ characters) or OIDC. Agents and Traefik use <code>TETHER_API_TOKEN</code> (32+ characters).
					Without a password Tether only answers on localhost.
				</p>
				<p>Pin the image to <code>0.7.0</code> until you have switched.</p>
			</Alert.Description>
		</Alert.Root>
		<main class="mx-auto flex w-full max-w-5xl flex-1 flex-col">
			{@render children()}
		</main>
		<AppFooter />
	</div>
{:else}
	<Login />
{/if}
