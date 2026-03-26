<script lang="ts">
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import FolderSyncIcon from '@lucide/svelte/icons/folder-sync';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import LucideIcon from '$lib/components/lucide-icon.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { toast } from 'svelte-sonner';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { Button } from '$lib/components/ui/button';
  import { fade } from 'svelte/transition';

  let {
    item
  }: {
    item: {
      id: string;
      title: string;
      icon?: string;
      books?: number;
    };
  } = $props();

  const sidebar = useSidebar();

  import { booksState } from '$lib/api/books.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';

  let confirmOpen = $state(false);
  let isDeleting = $state(false);
  let isScanning = $state(false);
  let dropdownOpen = $state(false);

  async function handleScan() {
    isScanning = true;
    try {
      const token = localStorage.getItem('bearer_token') || '';
      const res = await fetch(`/api/libraries/${item.id}/scan`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error(await res.text());
      const result = await res.json();
      toast.success(`Scan complete: ${result.added} added, ${result.removed} removed.`);
      booksState.invalidate(item.id);
      await Promise.all([
        librariesState.fetchAll(),
        shelvesState.fetchAll(),
        booksState.fetchAll()
      ]);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : String(e));
    } finally {
      isScanning = false;
      dropdownOpen = false;
    }
  }

  async function handleDelete() {
    isDeleting = true;
    try {
      await librariesState.delete(item.id);
      toast.success(`Library "${item.title}" deleted.`);
      confirmOpen = false;
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : String(e));
    } finally {
      isDeleting = false;
    }
  }
</script>

<div transition:fade>
  <Sidebar.MenuItem>
    <Sidebar.MenuButton tooltipContent={item.title}>
      {#snippet child({ props })}
        <a href={'/library/' + item.id} {...props}>
          {#if sidebar.state === 'collapsed'}
            {#if item.icon}
              <LucideIcon name={item.icon} />
            {:else}
              <span>{item.title.slice(0, 2)}</span>
            {/if}
          {:else}
            {#if item.icon}
              <LucideIcon name={item.icon} />
            {/if}
            <span>{item.title}</span>
          {/if}
        </a>
      {/snippet}
    </Sidebar.MenuButton>

    <DropdownMenu.Root bind:open={dropdownOpen}>
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <Sidebar.MenuAction
            class="peer/action top-1/2! z-10 aspect-auto size-7 -translate-y-1/2! bg-transparent opacity-0 transition-opacity group-hover/menu-item:opacity-100"
            {...props}
          >
            <EllipsisIcon />
            <span class="sr-only">More</span>
          </Sidebar.MenuAction>
        {/snippet}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content class="w-48 rounded-lg" align={sidebar.isMobile ? 'end' : 'start'}>
        <DropdownMenu.Item onclick={handleScan} disabled={isScanning}>
          <FolderSyncIcon class="text-muted-foreground" />
          <span>{isScanning ? 'Scanning…' : 'Rescan Library'}</span>
        </DropdownMenu.Item>
        <DropdownMenu.Separator />
        <DropdownMenu.Item onclick={() => (confirmOpen = true)}>
          <Trash2Icon class="text-muted-foreground" />
          <span>Delete Library</span>
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <AlertDialog.Root bind:open={confirmOpen}>
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
          <AlertDialog.Description>
            This action cannot be undone. This will permanently delete the
            <strong>{item.title}</strong> library.
          </AlertDialog.Description>
        </AlertDialog.Header>
        <AlertDialog.Footer>
          <AlertDialog.Cancel disabled={isDeleting}>Cancel</AlertDialog.Cancel>
          <Button variant="destructive" onclick={handleDelete} disabled={isDeleting}>
            {isDeleting ? 'Deleting...' : 'Delete Library'}
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog.Root>

    <Sidebar.MenuBadge class="transition-opacity group-hover/menu-item:opacity-0">
      {item.books ?? 0}
    </Sidebar.MenuBadge>
  </Sidebar.MenuItem>
</div>
