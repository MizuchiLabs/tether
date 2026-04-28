<script lang="ts">
	import './layout.css';
	import { onMount } from 'svelte';
	import Logo from '$lib/assets/logo.svelte';
	import { Button } from '$lib/components/ui/button';
	import { LogOut, Moon, Sun } from '@lucide/svelte';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import Login from '$lib/components/Login.svelte';
	import { loggedIn } from '$lib/store.svelte';
	import { api } from '$lib/api';

	const { children } = $props();

	onMount(async () => {
		try {
			await api.envs();
			loggedIn.current = true;
		} catch (_) {
			loggedIn.current = false;
		}
	});
</script>

<ModeWatcher />
<Login />

{#if loggedIn.current}
	<div class="flex min-h-screen flex-col">
		<header class="sticky top-0 z-40 w-full border-b bg-background/80 backdrop-blur-md shadow-sm">
			<div class="container mx-auto flex h-14 items-center justify-between px-6 sm:px-0">
				<div class="flex items-center gap-3">
					<Logo class="size-6" />
					<h1 class="text-xl font-bold tracking-tight">Tether</h1>
				</div>

				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={toggleMode} class="relative">
						<Sun
							class="rotate-0 scale-100 transition-all duration-300 dark:-rotate-90 dark:scale-0"
						/>
						<Moon
							class="absolute rotate-90 scale-0 transition-all duration-300 dark:rotate-0 dark:scale-100"
						/>
						<span class="sr-only">Toggle theme</span>
					</Button>

					<Button variant="ghost" size="icon" onclick={api.logout} title="Log out">
						<LogOut />
					</Button>
				</div>
			</div>
		</header>

		<main class="flex flex-1 flex-col container mx-auto py-6 px-4 sm:px-0">
			{@render children()}
		</main>
	</div>
{/if}
