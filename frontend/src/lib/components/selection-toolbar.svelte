<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { buttonVariants } from '$lib/components/ui/button';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import XIcon from '@lucide/svelte/icons/x';
  import { toast } from 'svelte-sonner';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import SquareCheckBig from '@lucide/svelte/icons/square-check-big';

  let {
    selectedIds,
    ondeselect,
    onselectall,
    books
  }: {
    selectedIds: Set<string>;
    ondeselect: () => void;
    onselectall: () => void;
    books: Book[];
  } = $props();

  let deleteOpen = $state(false);
  let deleteFile = $state(false);
  let deleting = $state(false);

  let count = $derived(selectedIds.size);

  async function confirmDelete() {
    deleting = true;
    const ids = [...selectedIds];
    let failed = 0;
    for (const id of ids) {
      try {
        await booksState.delete(id, deleteFile);
      } catch {
        failed++;
      }
    }
    deleting = false;
    deleteOpen = false;
    ondeselect();
    if (failed > 0) {
      toast.error(`Failed to delete ${failed} book${failed > 1 ? 's' : ''}.`);
    } else {
      toast.success(`Deleted ${ids.length} book${ids.length > 1 ? 's' : ''}.`);
    }
  }

  let selectedTitles = $derived(
    books
      .filter((b) => selectedIds.has(b.id))
      .map((b) => b.title)
      .slice(0, 3)
  );
</script>

{#if count > 0}
  <div
    class="fixed bottom-6 left-1/2 z-50 flex w-3xl -translate-x-1/2 items-center justify-between rounded-lg border bg-card px-3 py-2 shadow-lg"
  >
    <div class="flex items-center">
      <span class="text-sm font-medium">{count} selected</span>
    </div>

    <div class="flex items-center gap-x-2">
      <div class="h-8 w-px bg-border"></div>
      <!-- Select All -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={onselectall}
          >
            <SquareCheckBig class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Select All</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>
      <!-- Unselect All -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={ondeselect}
          >
            <XIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Clear Selection</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <div class="h-8 w-px bg-border"></div>

      <!-- Delete Selected -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'destructive', size: 'icon' })}
            onclick={() => (deleteOpen = true)}
          >
            <TrashIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Delete Selected Books</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>
    </div>
    <div></div>
  </div>
{/if}

<AlertDialog.Root bind:open={deleteOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete {count} book{count > 1 ? 's' : ''}?</AlertDialog.Title>
      <AlertDialog.Description>
        {#if selectedTitles.length < count}
          "{selectedTitles.join('", "')}" and {count - selectedTitles.length} more will be removed from
          your library. This action cannot be undone.
        {:else}
          "{selectedTitles.join('", "')}" will be removed from your library. This action cannot be
          undone.
        {/if}
      </AlertDialog.Description>
    </AlertDialog.Header>

    <label class="flex cursor-pointer items-center gap-2 text-sm">
      <input type="checkbox" bind:checked={deleteFile} class="size-4 accent-primary" />
      Also delete the files from disk
    </label>

    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
