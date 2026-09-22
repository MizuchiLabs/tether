<script lang="ts">
	import Login from '$lib/components/Login.svelte';
	import AppFooter from '$lib/components/nav/AppFooter.svelte';
	import AppHeader from '$lib/components/nav/AppHeader.svelte';
	import { Toaster } from '$lib/components/ui/sonner';
	import { loggedIn } from '$lib/store.svelte';
	import { ModeWatcher } from 'mode-watcher';
	import './layout.css';

	const { children } = $props();
</script>

<ModeWatcher />
<Toaster richColors />
<Login />

{#if loggedIn.current}
	<div class="flex min-h-screen flex-col">
		<AppHeader />
		<main class="z-10 mb-12 flex-1">
			{@render children?.()}
		</main>
		<AppFooter />
	</div>
{/if}
