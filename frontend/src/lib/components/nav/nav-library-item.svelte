<script lang="ts">
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import ForwardIcon from '@lucide/svelte/icons/forward';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import type { Component } from 'svelte';

  let {
    item
  }: {
    item: {
      title: string;
      url: string;
      icon?: Component;
      books?: number;
    };
  } = $props();

  const sidebar = useSidebar();
</script>

<Sidebar.MenuItem>
  <Sidebar.MenuButton tooltipContent={item.title}>
    {#snippet child({ props })}
      <a href={item.url} {...props}>
        {#if sidebar.state === 'collapsed'}
          {#if item.icon}
            <item.icon />
          {:else}
            <span>{item.title.slice(0, 2)}</span>
          {/if}
        {:else}
          {#if item.icon}
            <item.icon />
          {/if}
          <span>{item.title}</span>
        {/if}
      </a>
    {/snippet}
  </Sidebar.MenuButton>

  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Sidebar.MenuAction
          class="peer/action z-10 aspect-auto size-6 bg-transparent opacity-0 hover:opacity-100 focus-visible:opacity-100 data-[state=open]:opacity-100"
          {...props}
        >
          <EllipsisIcon />
          <span class="sr-only">More</span>
        </Sidebar.MenuAction>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-48 rounded-lg" align={sidebar.isMobile ? 'end' : 'start'}>
      <DropdownMenu.Item>
        <FolderIcon class="text-muted-foreground" />
        <span>View Project</span>
      </DropdownMenu.Item>
      <DropdownMenu.Item>
        <ForwardIcon class="text-muted-foreground" />
        <span>Share Project</span>
      </DropdownMenu.Item>
      <DropdownMenu.Separator />
      <DropdownMenu.Item>
        <Trash2Icon class="text-muted-foreground" />
        <span>Delete Project</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>

  <Sidebar.MenuBadge
    class="transition-opacity peer-hover/action:opacity-0 peer-focus-visible/action:opacity-0 peer-data-[state=open]/action:opacity-0"
  >
    {item.books ?? 0}
  </Sidebar.MenuBadge>
</Sidebar.MenuItem>
