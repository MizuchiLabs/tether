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
		await api.envs();
	});
</script>

<ModeWatcher />
<Login />

{#if loggedIn.current}
	<div class="relative flex min-h-screen flex-col bg-background">
		<header
			class="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60"
		>
			<div class="container mx-auto flex h-14 max-w-screen-2xl items-center px-4 sm:px-8">
				<div class="mr-4 flex items-center gap-6 md:mr-6">
					<a href="/" class="flex items-center gap-2 transition-opacity hover:opacity-80">
						<Logo class="h-6 w-6" />
						<span class="hidden font-bold tracking-tight sm:inline-block">Tether</span>
					</a>
				</div>

				<div class="flex flex-1 items-center justify-end gap-2">
					<nav class="flex items-center gap-1">
						<Button variant="ghost" size="icon" onclick={toggleMode}>
							<Sun class="dark:hidden" />
							<Moon class="hidden dark:block" />
							<span class="sr-only">Toggle theme</span>
						</Button>

						<Button
							variant="ghost"
							onclick={api.logout}
							class="gap-2 px-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive md:px-3"
							title="Log out"
						>
							<LogOut />
							<span class="hidden text-sm md:inline-block">Log out</span>
						</Button>
					</nav>
				</div>
			</div>
		</header>

		<main class="flex-1">
			<div class="container mx-auto max-w-screen-2xl px-4 py-6 sm:px-8 md:py-8">
				{@render children()}
			</div>
		</main>
	</div>
{/if}
