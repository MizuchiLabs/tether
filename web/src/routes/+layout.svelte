<script lang="ts">
	import { api } from '#lib/api.js';
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
		{#await api.v08Ready() then ready}
			{#if !ready}
				<Alert.Root class="mx-auto mb-6 max-w-5xl">
					<TriangleAlertIcon />
					<Alert.Title>Sign in changes in v0.8.0 on January 13</Alert.Title>
					<Alert.Description>
						<p>
							<code>TETHER_TOKEN</code> goes away. Set <code>TETHER_API_TOKEN</code> (32+
							characters) for agents and Traefik, and <code>TETHER_PASSWORD</code> (12+ characters) or
							OIDC for the web UI. This version already reads them, so you can switch now. Until v0.8.0
							the web UI still signs in with the API token.
						</p>
						<p>
							<a class="underline" href="https://github.com/MizuchiLabs/tether#moving-to-v080">
								How to move over
							</a>
						</p>
					</Alert.Description>
				</Alert.Root>
			{/if}
		{/await}
		<main class="mx-auto flex w-full max-w-5xl flex-1 flex-col">
			{@render children()}
		</main>
		<AppFooter />
	</div>
{:else}
	<Login />
{/if}
