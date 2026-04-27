<script lang="ts">
	import './layout.css';
	import { onMount } from 'svelte';
	import { authState } from '$lib/auth.svelte';
	import Logo from '$lib/assets/logo.svelte';
	import { Button } from '$lib/ui/button';
	import { LogOut, Moon, Sun } from '@lucide/svelte';
	import { ModeWatcher, toggleMode, mode } from 'mode-watcher';

	const { children } = $props();

	onMount(async () => {
		// Initial check
		try {
			const res = await fetch('/envs', {
				headers: authState.token ? { Authorization: `Bearer ${authState.token}` } : {}
			});
			if (res.ok) {
				authState.isAuthed = true;
			}
		} catch (e) {
			// Ignore initial fail, Login component will handle it
		}
	});

	function logout() {
		authState.setToken('');
		authState.isAuthed = false;
	}
</script>

<ModeWatcher />

<div class="flex min-h-screen flex-col">
	<header
		class="sticky top-0 z-10 flex h-16 items-center justify-between border-b border-accent-foreground px-6 bg-card shadow-sm"
	>
		<div class="flex items-center gap-3">
			<Logo class="h-6 w-6" />
			<h1 class="text-xl font-bold tracking-tight">Tether</h1>
		</div>
		<div class="flex items-center gap-3">
			<Button variant="ghost" size="icon" onclick={toggleMode}>
				{#if mode.current === 'light'}
					<Sun />
				{:else}
					<Moon />
				{/if}
			</Button>

			{#if authState.isAuthed}
				<Button variant="ghost" size="icon" onclick={logout}>
					<LogOut />
				</Button>
			{/if}
		</div>
	</header>

	<main class="flex flex-1 flex-col p-6">
		{@render children()}
	</main>
</div>
