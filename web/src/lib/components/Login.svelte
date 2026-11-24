<script lang="ts">
	import { api, ApiError, type Session } from '#lib/api.js';
	import Logo from '#lib/assets/logo.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Card from '#lib/components/ui/card/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import * as InputGroup from '#lib/components/ui/input-group/index.js';
	import { Spinner } from '#lib/components/ui/spinner/index.js';
	import { EyeIcon, EyeOffIcon } from '@lucide/svelte';

	const { mode }: { mode: Session['mode'] } = $props();

	let password = $state('');
	let showPassword = $state(false);
	let loading = $state(false);
	let error = $state<string | null>(null);

	async function login(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		error = null;
		try {
			await api.login(password);
			password = '';
		} catch (err) {
			if (err instanceof ApiError) {
				if (err.status === 401) error = 'Wrong password, try again.';
				else if (err.status === 429) error = 'Too many tries, wait a minute.';
				else error = err.message;
			} else {
				error = "Can't reach Tether right now.";
			}
		} finally {
			loading = false;
		}
	}
</script>

<main class="flex min-h-svh items-center justify-center px-4 py-16">
	<Card.Root class="w-full max-w-sm">
		<Card.Header class="justify-items-center text-center">
			<Logo class="mb-2 size-8" />
			<Card.Title>Sign in to Tether</Card.Title>
			<Card.Description>
				{#if mode === 'oidc'}
					Use your account to open Tether.
				{:else}
					Use the password from <code>TETHER_PASSWORD</code>.
				{/if}
			</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if mode === 'oidc'}
				<!-- A page load, the browser goes to the provider and comes back. -->
				<Button href="/api/session/oidc" class="w-full" data-sveltekit-reload>Sign in</Button>
			{:else}
				<form onsubmit={login}>
					<Field.FieldGroup>
						<Field.Field data-invalid={!!error || undefined}>
							<Field.FieldLabel for="password">Password</Field.FieldLabel>
							<InputGroup.Root>
								<InputGroup.Input
									id="password"
									type={showPassword ? 'text' : 'password'}
									autocomplete="current-password"
									autofocus
									required
									bind:value={password}
									aria-invalid={!!error || undefined}
								/>
								<InputGroup.Addon align="inline-end">
									<InputGroup.Button
										size="icon-xs"
										aria-label={showPassword ? 'Hide password' : 'Show password'}
										title={showPassword ? 'Hide password' : 'Show password'}
										onclick={() => (showPassword = !showPassword)}
									>
										{#if showPassword}
											<EyeOffIcon />
										{:else}
											<EyeIcon />
										{/if}
									</InputGroup.Button>
								</InputGroup.Addon>
							</InputGroup.Root>
							{#if error}
								<Field.FieldError>{error}</Field.FieldError>
							{/if}
						</Field.Field>

						<Button type="submit" disabled={loading || !password}>
							{#if loading}
								<Spinner data-icon="inline-start" />
								Signing in...
							{:else}
								Sign in
							{/if}
						</Button>
					</Field.FieldGroup>
				</form>
			{/if}
		</Card.Content>
	</Card.Root>
</main>
