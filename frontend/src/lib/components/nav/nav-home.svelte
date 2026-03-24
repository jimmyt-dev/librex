<script lang="ts">
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import type { Component } from 'svelte';

  let {
    links
  }: {
    links: {
      title: string;
      url: string;
      icon: Component;
      count?: number;
    }[];
  } = $props();

  // Removed commented useSidebar
</script>

<Sidebar.Group>
  <Sidebar.GroupLabel class="text-sm text-foreground">Home</Sidebar.GroupLabel>
  <Sidebar.Menu>
    {#each links as item (item.title)}
      <Sidebar.MenuItem>
        <Sidebar.MenuButton tooltipContent={item.title}>
          {#snippet child({ props })}
            <a href={item.url} {...props}>
              <item.icon />
              <span>{item.title}</span>
            </a>
          {/snippet}
        </Sidebar.MenuButton>
        {#if item.count != null}
          <Sidebar.MenuBadge>{item.count}</Sidebar.MenuBadge>
        {/if}
      </Sidebar.MenuItem>
    {/each}
  </Sidebar.Menu>
</Sidebar.Group>
