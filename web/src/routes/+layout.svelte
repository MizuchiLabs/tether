<script lang="ts">
	import { api } from '$lib/api';
	import Logo from '$lib/assets/logo.svelte';
	import Login from '$lib/components/Login.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Toaster } from '$lib/components/ui/sonner';
	import { loggedIn } from '$lib/store.svelte';
	import { LogOut, Moon, Sun } from '@lucide/svelte';
	import { ModeWatcher, toggleMode } from 'mode-watcher';
	import './layout.css';

	const { children } = $props();
</script>

<ModeWatcher />
<Toaster richColors />
<Login />

{#if loggedIn.current}
	<div class="relative flex min-h-screen flex-col bg-background">
		<header
			class="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60"
		>
			<div class="container mx-auto flex h-14 max-w-screen-2xl items-center gap-6 px-4 sm:px-8">
				<a href="/" class="flex items-center gap-2 transition-opacity hover:opacity-80">
					<Logo class="size-6" />
					<span class="hidden font-bold tracking-tight sm:inline-block">Tether</span>
				</a>

				<div class="flex flex-1 items-center justify-end">
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
							<LogOut data-icon="inline-start" />
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
