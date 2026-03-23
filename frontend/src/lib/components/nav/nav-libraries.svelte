<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import NavLibraryItem from './nav-library-item.svelte';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import CreateLibrary from '../create-library.svelte';

  const sidebar = useSidebar();

  let userOpen = $state(true);
  let open = $derived(sidebar.state === 'collapsed' || userOpen);

  let {
    links,
    onAdd
  }: {
    links: {
      id: string;
      title: string;
      url: string;
      icon?: string;
      books: number;
    }[];
    onAdd?: (name: string, icon?: string) => Promise<void>;
  } = $props();
</script>

<Collapsible.Root {open} class="group/collapsible" onOpenChange={(value) => (userOpen = value)}>
  <Sidebar.Group>
    <Sidebar.GroupLabel
      class="flex w-full items-center justify-between text-sm text-foreground group-data-[collapsible=icon]:pointer-events-none"
    >
      <span class="flex items-center justify-center gap-1">
        Libraries
        {#if onAdd}
          <CreateLibrary {onAdd} />
        {/if}
      </span>
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
