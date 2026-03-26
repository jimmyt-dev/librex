<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { buttonVariants } from '$lib/components/ui/button';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import XIcon from '@lucide/svelte/icons/x';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import FolderSyncIcon from '@lucide/svelte/icons/folder-sync';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import PenLineIcon from '@lucide/svelte/icons/pen-line';
  import { toast } from 'svelte-sonner';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import SquareCheckBig from '@lucide/svelte/icons/square-check-big';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import BookBulkEditSheet from '$lib/components/book-bulk-edit-sheet.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { Label } from '$lib/components/ui/label';

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
  let moving = $state(false);
  let bulkEditOpen = $state(false);

  let count = $derived(selectedIds.size);

  async function confirmDelete() {
    deleting = true;
    const ids = [...selectedIds];
    let failed = 0;
    for (const id of ids) {
      try {
        await booksState.delete(id, deleteFile);
        shelvesState.removeBook(id);
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
    librariesState.fetchAll();
    shelvesState.fetchAll();
  }

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

  async function moveBooks() {
    moving = true;
    const ids = [...selectedIds];
    try {
      const res = await fetch('/api/books/move', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${getToken()}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ bookIds: ids })
      });
      if (!res.ok) throw new Error('Failed to move books');
      const results: { bookId: string; error?: string }[] = await res.json();
      const failed = results.filter((r) => r.error);
      if (failed.length > 0) {
        toast.error(`Failed to move ${failed.length} book${failed.length > 1 ? 's' : ''}.`);
      } else {
        toast.success(`Moved ${ids.length} book${ids.length > 1 ? 's' : ''} to naming pattern.`);
      }
      // Refresh book data
      booksState.fetchAll();
      for (const b of books) {
        if (ids.includes(b.id)) {
          booksState.fetchForLibrary(b.libraryId);
        }
      }
    } catch {
      toast.error('Failed to move books.');
    } finally {
      moving = false;
    }
  }

  let selectedTitles = $derived(
    books
      .filter((b) => selectedIds.has(b.id))
      .map((b) => b.metadata.title)
      .slice(0, 3)
  );
</script>

{#if count > 0}
  <div
    class="fixed bottom-6 left-1/2 z-50 flex w-[calc(100vw-2rem)] -translate-x-1/2 items-center justify-between rounded-lg border bg-card px-3 py-2 shadow-lg sm:w-xl md:w-2xl lg:w-3xl"
  >
    <div class="flex items-center">
      <span class="text-sm font-medium">{count} selected</span>
    </div>

    <div class="flex items-center gap-x-2">
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

      <!-- Add to Shelf -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={() => shelfAssignState.openFor([...selectedIds])}
          >
            <LibraryBigIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Shelves</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <!-- Bulk Edit -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={() => (bulkEditOpen = true)}
          >
            <PencilIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Bulk Edit</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <!-- Edit Each -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={() => bookEditState.openQueue(books.filter((b) => selectedIds.has(b.id)))}
          >
            <PenLineIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Edit Each</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <!-- Move to Naming Pattern -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={moveBooks}
            disabled={moving}
          >
            <FolderSyncIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Move to Naming Pattern</Tooltip.Content>
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

<BookBulkEditSheet bind:open={bulkEditOpen} {selectedIds} {books} />

<AlertDialog.Root
  bind:open={deleteOpen}
  onOpenChange={(o) => {
    if (!o) deleteFile = false;
  }}
>
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

    <Label class="flex cursor-pointer items-center gap-2 text-sm">
      <Checkbox bind:checked={deleteFile} />
      Also delete the files from disk
    </Label>

    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
