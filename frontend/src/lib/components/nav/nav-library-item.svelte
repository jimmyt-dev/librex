<script lang="ts">
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import ForwardIcon from '@lucide/svelte/icons/forward';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import LucideIcon from '$lib/components/ui/lucide-icon.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { toast } from 'svelte-sonner';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { Button } from '$lib/components/ui/button';

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

  let confirmOpen = $state(false);
  let isDeleting = $state(false);

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
      <DropdownMenu.Item onclick={() => (confirmOpen = true)}>
        <Trash2Icon class="text-muted-foreground" />
        <span>Delete Project</span>
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

  <Sidebar.MenuBadge
    class="transition-opacity peer-hover/action:opacity-0 peer-focus-visible/action:opacity-0 peer-data-[state=open]/action:opacity-0"
  >
    {item.books ?? 0}
  </Sidebar.MenuBadge>
</Sidebar.MenuItem>
