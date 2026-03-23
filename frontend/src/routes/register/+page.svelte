<script lang="ts">
  import { enhance } from '$app/forms';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { ActionData } from './$types';

  let { form }: { form: ActionData } = $props();
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
      <form method="post" use:enhance class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <label for="name" class="text-sm font-medium text-foreground">Name</label>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Jane Smith"
            autocomplete="name"
            required
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="email" class="text-sm font-medium text-foreground">Email</label>
          <Input
            id="email"
            name="email"
            type="email"
            placeholder="you@example.com"
            autocomplete="email"
            required
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <label for="password" class="text-sm font-medium text-foreground">Password</label>
          <Input
            id="password"
            name="password"
            type="password"
            placeholder="••••••••"
            autocomplete="new-password"
            required
          />
        </div>

        {#if form?.message}
          <p class="text-sm text-destructive">{form.message}</p>
        {/if}

        <Button type="submit" class="mt-1 w-full">Create account</Button>
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
