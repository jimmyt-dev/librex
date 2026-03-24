<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import BookIcon from '@lucide/svelte/icons/book';
  import EllipsisVerticalIcon from '@lucide/svelte/icons/ellipsis-vertical';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import { toast } from 'svelte-sonner';
  import { Checkbox } from '$lib/components/ui/checkbox';

  let {
    book,
    selected = false,
    selectMode = false,
    checkboxes = true,
    onselect
  }: {
    book: Book;
    selected?: boolean;
    selectMode?: boolean;
    checkboxes?: boolean;
    onselect?: (id: string, selected: boolean, shiftKey: boolean) => void;
  } = $props();

  let lastClickShift = false;
  let deleteOpen = $state(false);
  let deleteFile = $state(false);
  let deleting = $state(false);
  let dropdownOpen = $state(false);

  function handleCardClick(e: MouseEvent) {
    if (selectMode) {
      onselect?.(book.id, !selected, e.shiftKey);
    }
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await booksState.delete(book.id, deleteFile);
      shelvesState.removeBook(book.id);
      toast.success(`"${book.title}" deleted.`);
      deleteOpen = false;
      librariesState.fetchAll();
      shelvesState.fetchAll();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete book.');
    } finally {
      deleting = false;
    }
  }
</script>

<div
  class="group relative cursor-pointer outline-none"
  role="button"
  tabindex="0"
  onclick={(e) => handleCardClick(e)}
  onkeydown={(e) => e.key === 'Enter' && handleCardClick(new MouseEvent('click'))}
>
  {#if checkboxes}
    <button
      class="absolute top-1.5 left-1.5 z-10 opacity-0 transition-opacity group-hover:opacity-100"
      class:opacity-100={selected}
      onclick={(e) => e.stopPropagation()}
    >
      <Checkbox
        checked={selected}
        onCheckedChange={(v) => onselect?.(book.id, !!v, lastClickShift)}
        onclick={(e) => (lastClickShift = e.shiftKey)}
        class="border-2 border-white bg-white/70 shadow-md drop-shadow-sm backdrop-blur-sm data-[state=checked]:border-primary data-[state=checked]:bg-primary"
        aria-label="Select {book.title}"
      />
    </button>
  {/if}

  <div
    class="focus-none flex flex-col overflow-hidden rounded-md border bg-card text-card-foreground shadow-sm transition-all"
    class:ring-2={selected}
    class:ring-primary={selected}
  >
    <div class="relative aspect-2/3 w-full overflow-hidden bg-muted">
      {#if book.cover}
        <img
          src={book.cover}
          alt={book.title}
          class="h-full w-full object-cover transition duration-300 group-hover:scale-105"
        />
      {:else}
        <div class="flex h-full w-full items-center justify-center text-muted-foreground">
          <BookIcon class="size-8" />
        </div>
      {/if}
    </div>

    <div class="flex min-w-0 items-stretch">
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger class="min-w-0 flex-1 px-2 py-2 text-left">
            <p class="truncate text-xs leading-tight font-medium">{book.title}</p>
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content side="bottom">{book.title}</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <DropdownMenu.Root bind:open={dropdownOpen}>
        <DropdownMenu.Trigger
          onclick={(e) => e.stopPropagation()}
          class="flex shrink-0 items-center justify-around rounded-none px-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <EllipsisVerticalIcon class="size-4" />
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content align="start" class="w-40">
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                bookEditState.openFor(book);
              }}
            >
              <PencilIcon class="size-3.5" />
              Edit
            </DropdownMenu.Item>
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                shelfAssignState.openFor([book.id]);
              }}
            >
              <LibraryBigIcon class="size-3.5" />
              Shelves
            </DropdownMenu.Item>
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              class="text-destructive focus:text-destructive"
              onclick={(e) => {
                e.stopPropagation();
                deleteOpen = true;
              }}
            >
              <TrashIcon class="size-3.5" />
              Delete
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
    </div>
  </div>
</div>

<AlertDialog.Root
  bind:open={deleteOpen}
  onOpenChange={(o) => {
    if (!o) deleteFile = false;
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete "{book.title}"?</AlertDialog.Title>
      <AlertDialog.Description>
        This will remove the book from your library. This action cannot be undone.
      </AlertDialog.Description>
    </AlertDialog.Header>
    <label class="flex cursor-pointer items-center gap-2 text-sm">
      <Checkbox bind:checked={deleteFile} />
      Also delete the file from disk
    </label>
    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
