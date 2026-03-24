<script lang="ts">
  import type { Component } from 'svelte';
  import * as Popover from '$lib/components/ui/popover';
  import { ScrollArea } from '$lib/components/ui/scroll-area';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import LucideIcon from '$lib/components/lucide-icon.svelte';
  import XIcon from '@lucide/svelte/icons/x';

  let {
    value = $bindable(''),
    placeholder = 'Select icon...'
  }: { value?: string; placeholder?: string } = $props();

  const icons = import.meta.glob<{ default: Component }>(
    '/node_modules/@lucide/svelte/dist/icons/*.svelte',
    { eager: false }
  );

  const allIconNames = Object.keys(icons)
    .map((path) =>
      path.replace('/node_modules/@lucide/svelte/dist/icons/', '').replace('.svelte', '')
    )
    .sort();

  let search = $state('');
  let open = $state(false);

  let filtered = $derived(
    search ? allIconNames.filter((name) => name.includes(search.toLowerCase())) : allIconNames
  );

  // Only render a limited number to avoid performance issues
  let visible = $derived(filtered.slice(0, 100));

  function select(name: string) {
    value = name;
    open = false;
    search = '';
  }

  function clear() {
    value = '';
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        variant="outline"
        {...props}
        class="w-full justify-start gap-2 font-normal {!value ? 'text-muted-foreground' : ''}"
      >
        {#if value}
          <LucideIcon name={value} class="size-4" />
          <span class="truncate">{value}</span>
          <button
            type="button"
            class="ml-auto hover:text-foreground"
            onclick={(e) => {
              e.stopPropagation();
              clear();
            }}
          >
            <XIcon class="size-3" />
          </button>
        {:else}
          {placeholder}
        {/if}
      </Button>
    {/snippet}
  </Popover.Trigger>
  <Popover.Content class="w-80 p-0" align="start">
    <div class="p-2">
      <Input bind:value={search} placeholder="Search icons..." class="h-8" />
    </div>
    <ScrollArea class="h-64">
      <div class="grid grid-cols-6 gap-1 p-2 pt-0">
        {#each visible as name (name)}
          <button
            type="button"
            class="flex items-center justify-center rounded-md p-2 hover:cursor-pointer hover:bg-accent {value ===
            name
              ? 'bg-accent ring-1 ring-ring'
              : ''}"
            title={name}
            onclick={() => select(name)}
          >
            <LucideIcon {name} class="size-4" />
          </button>
        {/each}
      </div>
      {#if filtered.length > 100}
        <p class="px-2 pb-2 text-center text-xs text-muted-foreground">
          Showing 100 of {filtered.length} icons. Type to narrow results.
        </p>
      {/if}
      {#if filtered.length === 0}
        <p class="px-2 pb-2 text-center text-xs text-muted-foreground">No icons found.</p>
      {/if}
    </ScrollArea>
  </Popover.Content>
</Popover.Root>
