<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import NavLibraryItem from './nav-library-item.svelte';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import CreateLibrary from '../create-library.svelte';
  import PlusIcon from '@lucide/svelte/icons/plus';

  const sidebar = useSidebar();

  let createOpen = $state(false);
  let userOpen = $state(true);
  let open = $derived(sidebar.state === 'collapsed' || userOpen);

  let {
    links
  }: {
    links: {
      id: string;
      title: string;
      icon?: string;
      books: number;
    }[];
  } = $props();
</script>

<Collapsible.Root {open} class="group/collapsible" onOpenChange={(value) => (userOpen = value)}>
  <Sidebar.Group>
    <Sidebar.GroupLabel
      class="flex w-full items-center justify-between text-sm text-foreground group-data-[collapsible=icon]:pointer-events-none"
    >
      <span class="flex items-center justify-center gap-1">
        Libraries
        <button
          type="button"
          class="flex items-center justify-center hover:cursor-pointer hover:text-foreground/80"
          onclick={() => (createOpen = true)}
          aria-label="Create library"
        >
          <PlusIcon class="size-4" />
        </button>
      </span>
      <CreateLibrary bind:open={createOpen} />
      {#if links.length > 0}
        <Collapsible.Trigger>
          <ChevronRightIcon
            class="ms-auto size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90 hover:cursor-pointer"
          />
        </Collapsible.Trigger>
      {/if}
    </Sidebar.GroupLabel>
    <Collapsible.Content>
      <Sidebar.Menu>
        {#each links as item (item.id)}
          <NavLibraryItem {item} />
        {/each}
      </Sidebar.Menu>
    </Collapsible.Content>
  </Sidebar.Group>
</Collapsible.Root>
