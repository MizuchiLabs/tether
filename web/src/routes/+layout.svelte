<script lang="ts">
	import Login from '#lib/components/Login.svelte';
	import AppFooter from '#lib/components/nav/AppFooter.svelte';
	import AppHeader from '#lib/components/nav/AppHeader.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import { loggedIn } from '#lib/store.svelte.js';
	import { ModeWatcher } from 'mode-watcher';
	import './layout.css';

	const { data, children } = $props();
</script>

<!-- Same values as --background in layout.css. -->
<ModeWatcher themeColors={{ dark: '#181825', light: '#eff1f5' }} />
<Toaster richColors />

{#if loggedIn.current}
	<div class="flex min-h-svh flex-col px-4">
		<AppHeader canSignOut={data.mode !== 'none'} />
		<main class="mx-auto flex w-full max-w-5xl flex-1 flex-col">
			{@render children()}
		</main>
		<AppFooter />
	</div>
{:else}
	<Login mode={data.mode} />
{/if}
