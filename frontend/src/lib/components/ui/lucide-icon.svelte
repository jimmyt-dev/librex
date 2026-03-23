<script lang="ts">
  import type { Component } from 'svelte';

  let { name, class: className = '' }: { name: string; class?: string } = $props();

  const icons = import.meta.glob<{ default: Component }>(
    '/node_modules/@lucide/svelte/dist/icons/*.svelte',
    { eager: false }
  );

  let icon = $state<Component | null>(null);

  $effect(() => {
    icon = null;
    const path = `/node_modules/@lucide/svelte/dist/icons/${name}.svelte`;
    const loader = icons[path];
    if (loader) {
      loader().then((mod) => {
        icon = mod.default;
      });
    }
  });
</script>

{#if icon}
  <svelte:component this={icon} class={className} />
{/if}
