<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import type { Component } from 'svelte';
  import NavShelfItem from './nav-shelf-item.svelte';
  import { useSidebar } from '$lib/components/ui/sidebar';

  const sidebar = useSidebar();
  $effect(() => {
    console.log(sidebar.state);
    console.log(open);
  });
  let userOpen = $state(true);
  let open = $derived(sidebar.state === 'collapsed' || userOpen);

  let {
    links
  }: {
    links: {
      title: string;
      url: string;
      icon?: Component;
      books: number;
    }[];
  } = $props();
</script>

<Collapsible.Root {open} class="group/collapsible" onOpenChange={(value) => (userOpen = value)}>
  <Sidebar.Group>
    <Sidebar.GroupLabel class="flex w-full items-center justify-between text-sm text-foreground">
      Shelves
      <Collapsible.Trigger>
        <ChevronRightIcon
          class="ms-auto size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90 hover:cursor-pointer"
        />
      </Collapsible.Trigger>
    </Sidebar.GroupLabel>
    <Collapsible.Content>
      <Sidebar.Menu>
        {#each links as item (item.title)}
          <NavShelfItem {item} />
        {/each}
      </Sidebar.Menu>
    </Collapsible.Content>
  </Sidebar.Group>
</Collapsible.Root>
