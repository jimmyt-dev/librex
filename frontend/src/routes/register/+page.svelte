<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { authClient } from '$lib/auth-client';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';

  let name = $state('');
  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;

    await authClient.signUp.email(
      { name, email, password },
      {
        onSuccess: async () => {
          await invalidateAll();
          goto('/');
        },
        onError: (ctx) => {
          error = ctx.error.message || 'Registration failed';
          loading = false;
        }
      }
    );
  }
</script>

<svelte:head>
  <title>Create account</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4">
  <div class="w-full max-w-sm">
    <div class="mb-8 text-center">
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">Create an account</h1>
      <p class="mt-1 text-sm text-muted-foreground">Get started by filling in the details below</p>
    </div>

    <div class="rounded-xl border border-border bg-card p-6 shadow-sm">
      <form onsubmit={handleSubmit} class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <label for="name" class="text-sm font-medium text-foreground">Name</label>
          <Input
            id="name"
            type="text"
            placeholder="Jane Smith"
            autocomplete="name"
            required
            bind:value={name}
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="email" class="text-sm font-medium text-foreground">Email</label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            autocomplete="email"
            required
            bind:value={email}
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="password" class="text-sm font-medium text-foreground">Password</label>
          <Input
            id="password"
            type="password"
            placeholder="••••••••"
            autocomplete="new-password"
            required
            bind:value={password}
          />
        </div>

        {#if error}
          <p class="text-sm text-destructive">{error}</p>
        {/if}

        <Button type="submit" class="mt-1 w-full" disabled={loading}>
          {loading ? 'Creating account...' : 'Create account'}
        </Button>
      </form>
    </div>

    <p class="mt-4 text-center text-sm text-muted-foreground">
      Already have an account?
      <a href="/login" class="font-medium text-foreground underline-offset-4 hover:underline">
        Sign in
      </a>
    </p>
  </div>
</div>
