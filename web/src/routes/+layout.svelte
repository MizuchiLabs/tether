<script lang="ts">
	import Login from '#lib/components/Login.svelte';
	import AppFooter from '#lib/components/nav/AppFooter.svelte';
	import AppHeader from '#lib/components/nav/AppHeader.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import { loggedIn } from '#lib/store.svelte.js';
	import { ModeWatcher } from 'mode-watcher';
	import './layout.css';

	const { children } = $props();
</script>

<ModeWatcher />
<Toaster richColors />

{#if loggedIn.current}
	<div class="flex min-h-screen flex-col px-4">
		<AppHeader />
		<main class="mx-auto flex w-full max-w-5xl flex-1 flex-col">
			{@render children()}
		</main>
		<AppFooter />
	</div>
{:else}
	<Login />
{/if}
