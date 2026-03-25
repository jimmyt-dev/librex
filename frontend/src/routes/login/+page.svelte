<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { authClient } from '$lib/auth-client';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;

    await authClient.signIn.email(
      { email, password },
      {
        onSuccess: async () => {
          await invalidateAll();
          goto('/');
        },
        onError: (ctx) => {
          error = ctx.error.message || 'Invalid email or password';
          loading = false;
        }
      }
    );
  }
</script>

<svelte:head>
  <title>Sign in</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4">
  <div class="w-full max-w-sm">
    <div class="mb-8 text-center">
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">Welcome back</h1>
      <p class="mt-1 text-sm text-muted-foreground">Sign in to your account to continue</p>
    </div>

    <div class="rounded-xl border border-border bg-card p-6 shadow-sm">
      <form onsubmit={handleSubmit} class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <Label for="email" class="text-sm font-medium text-foreground">Email</Label>
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
          <Label for="password" class="text-sm font-medium text-foreground">Password</Label>
          <Input
            id="password"
            type="password"
            placeholder="••••••••"
            autocomplete="current-password"
            required
            bind:value={password}
          />
        </div>

        {#if error}
          <p class="text-sm text-destructive">{error}</p>
        {/if}

        <Button type="submit" class="mt-1 w-full" disabled={loading}>
          {loading ? 'Signing in...' : 'Sign in'}
        </Button>
      </form>
    </div>

    <p class="mt-4 text-center text-sm text-muted-foreground">
      Don't have an account?
      <a href="/register" class="font-medium text-foreground underline-offset-4 hover:underline">
        Create one
      </a>
    </p>
  </div>
</div>
