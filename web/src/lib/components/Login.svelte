<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Field from '$lib/components/ui/field';
	import * as InputGroup from '$lib/components/ui/input-group';
	import { Spinner } from '$lib/components/ui/spinner';
	import { Eye, EyeOff } from '@lucide/svelte';
	import { api } from '$lib/api';
	import Logo from '$lib/assets/logo.svelte';
	import { loggedIn } from '$lib/store.svelte';

	let secret = $state('');
	let showPassword = $state(false);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	async function handleLogin(e: SubmitEvent) {
		e.preventDefault();

		isLoading = true;
		error = null;
		try {
			await api.login(secret);
			secret = '';
		} catch (err: any) {
			error = err.message || 'Failed to sign in. Please check your token.';
		} finally {
			isLoading = false;
		}
	}
</script>

{#if !loggedIn.current}
	<section class="flex min-h-screen px-4 py-16 md:py-32">
		<form
			onsubmit={handleLogin}
			class="m-auto h-fit w-full max-w-sm overflow-hidden rounded-[calc(var(--radius)+.125rem)] border bg-muted shadow-md shadow-zinc-950/5 dark:[--color-muted:var(--color-zinc-900)]"
		>
			<div class="-m-px rounded-[calc(var(--radius)+.125rem)] border bg-card p-8 pb-6">
				<div class="text-center">
					<a href="/" aria-label="go home" class="mx-auto block w-fit">
						<Logo class="size-7" />
					</a>
					<h1 class="mt-4 mb-1 text-xl font-semibold">Authenticate with Tether</h1>
					<p class="text-sm">Enter your shared secret token to view agent configurations.</p>
				</div>

				<Field.FieldGroup class="mt-6">
					<Field.Field data-invalid={!!error || undefined}>
						<Field.FieldLabel for="pwd">Access Token</Field.FieldLabel>
						<InputGroup.Root>
							<InputGroup.Input
								id="pwd"
								type={showPassword ? 'text' : 'password'}
								bind:value={secret}
								aria-invalid={!!error || undefined}
							/>
							<InputGroup.Addon align="inline-end">
								<Button
									aria-label={showPassword ? 'Hide password' : 'Show password'}
									title={showPassword ? 'Hide password' : 'Show password'}
									variant="ghost"
									size="icon-sm"
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<EyeOff data-icon="inline-start" />
									{:else}
										<Eye data-icon="inline-start" />
									{/if}
								</Button>
							</InputGroup.Addon>
						</InputGroup.Root>
						{#if error}
							<Field.FieldDescription class="text-destructive">{error}</Field.FieldDescription>
						{/if}
					</Field.Field>

					<Button type="submit" disabled={isLoading} class="w-full">
						{#if isLoading}
							<Spinner data-icon="inline-start" />
							Verifying...
						{:else}
							Sign In
						{/if}
					</Button>
				</Field.FieldGroup>
			</div>
		</form>
	</section>
{/if}
