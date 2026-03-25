<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { toast } from 'svelte-sonner';
  import BookIcon from '@lucide/svelte/icons/book';
  import EllipsisVerticalIcon from '@lucide/svelte/icons/ellipsis-vertical';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import { Label } from '$lib/components/ui/label';

  let {
    books,
    selectedIds,
    selectMode = false,
    onselect
  }: {
    books: Book[];
    selectedIds: Set<string>;
    selectMode?: boolean;
    onselect?: (id: string, selected: boolean, shiftKey: boolean) => void;
  } = $props();

  let bookToDelete = $state<Book | null>(null);
  let deleteFile = $state(false);
  let deleting = $state(false);

  async function confirmDelete() {
    if (!bookToDelete) return;
    deleting = true;
    try {
      await booksState.delete(bookToDelete.id, deleteFile);
      shelvesState.removeBook(bookToDelete.id);
      toast.success(`"${bookToDelete.metadata.title}" deleted.`);
      bookToDelete = null;
      librariesState.fetchAll();
      shelvesState.fetchAll();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete book.');
    } finally {
      deleting = false;
    }
  }

  async function download(book: Book) {
    try {
      const token = localStorage.getItem('bearer_token') || '';
      const res = await fetch(`/api/books/${book.id}/download`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error();
      const blob = await res.blob();
      const match = (res.headers.get('Content-Disposition') || '').match(/filename="(.+?)"/);
      const filename = match ? match[1] : book.metadata.title;
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = filename;
      a.click();
      URL.revokeObjectURL(a.href);
    } catch {
      toast.error('Failed to download book.');
    }
  }

  function formatSeries(book: Book): string {
    const { seriesName, seriesNumber } = book.metadata;
    if (!seriesName) return '—';
    if (seriesNumber == null) return seriesName;
    const num = Number.isInteger(seriesNumber)
      ? seriesNumber
      : seriesNumber.toFixed(2).replace(/\.?0+$/, '');
    return `${seriesName} #${num}`;
  }

  function formatExt(filePath: string): string {
    return filePath.split('.').pop()?.toUpperCase() ?? '—';
  }

  let allSelected = $derived(books.length > 0 && books.every((b) => selectedIds.has(b.id)));

  function toggleAll() {
    if (allSelected) {
      for (const b of books) onselect?.(b.id, false, false);
    } else {
      for (const b of books) {
        if (!selectedIds.has(b.id)) onselect?.(b.id, true, false);
      }
    }
  }
</script>

<div class="rounded-lg border">
  <table class="w-full text-sm">
    <thead>
      <tr class="border-b bg-muted/50 text-xs">
        <th class="w-10 px-4 py-2.5 text-left">
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <span onclick={toggleAll}>
            <Checkbox checked={allSelected} />
          </span>
        </th>
        <th class="w-10 px-2 py-2.5"></th>
        <th class="px-4 py-2.5 text-left font-medium">Title</th>
        <th class="px-4 py-2.5 text-left font-medium">Author</th>
        <th class="px-4 py-2.5 text-left font-medium">Series</th>
        <th class="px-4 py-2.5 text-left font-medium">Rating</th>
        <th class="px-4 py-2.5 text-left font-medium">Format</th>
        <th class="px-4 py-2.5"></th>
      </tr>
    </thead>
    <tbody>
      {#each books as book (book.id)}
        <tr
          class="border-b transition-colors last:border-0 hover:bg-muted/30 {selectedIds.has(
            book.id
          )
            ? 'bg-muted/20'
            : ''} {selectMode ? 'cursor-pointer' : ''}"
          onclick={(e) => {
            if (selectMode) onselect?.(book.id, !selectedIds.has(book.id), e.shiftKey);
          }}
        >
          <td class="px-4 py-2" onclick={(e) => e.stopPropagation()}>
            <Checkbox
              checked={selectedIds.has(book.id)}
              onCheckedChange={(v) => onselect?.(book.id, !!v, false)}
            />
          </td>

          <td class="px-2 py-2">
            {#if book.metadata.coverPath}
              <img
                src={`/api/books/${book.id}/cover`}
                alt=""
                class="h-10 w-7 rounded object-cover shadow-sm"
              />
            {:else}
              <div
                class="flex h-10 w-7 items-center justify-center rounded bg-muted text-muted-foreground"
              >
                <BookIcon class="size-4" />
              </div>
            {/if}
          </td>

          <td class="max-w-xs px-4 py-2">
            <a
              href="/books/{book.id}"
              class="font-medium hover:underline"
              onclick={(e) => e.stopPropagation()}
            >
              {book.metadata.title}
            </a>
            {#if book.metadata.subtitle}
              <p class="truncate text-xs text-muted-foreground">{book.metadata.subtitle}</p>
            {/if}
          </td>

          <td class="px-4 py-2 text-muted-foreground">
            {book.authors.map((a) => a.name).join(', ') || '—'}
          </td>

          <td class="px-4 py-2 text-muted-foreground">
            {formatSeries(book)}
          </td>

          <td class="px-4 py-2">
            {#if book.metadata.rating}
              <span class="text-base leading-none text-yellow-400"
                >{'★'.repeat(book.metadata.rating)}<span class="text-muted-foreground/30"
                  >{'★'.repeat(5 - book.metadata.rating)}</span
                ></span
              >
            {:else}
              <span class="text-muted-foreground">—</span>
            {/if}
          </td>

          <td class="px-4 py-2 text-xs text-muted-foreground">
            {formatExt(book.filePath)}
          </td>

          <td class="px-4 py-2" onclick={(e) => e.stopPropagation()}>
            <DropdownMenu.Root>
              <DropdownMenu.Trigger
                class="flex items-center justify-center rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <EllipsisVerticalIcon class="size-4" />
              </DropdownMenu.Trigger>
              <DropdownMenu.Portal>
                <DropdownMenu.Content align="end" class="w-40">
                  <DropdownMenu.Item onclick={() => bookEditState.openFor(book)}>
                    <PencilIcon class="size-3.5" /> Edit
                  </DropdownMenu.Item>
                  <DropdownMenu.Item onclick={() => shelfAssignState.openFor([book.id])}>
                    <LibraryBigIcon class="size-3.5" /> Shelves
                  </DropdownMenu.Item>
                  <DropdownMenu.Item onclick={() => download(book)}>
                    <DownloadIcon class="size-3.5" /> Download
                  </DropdownMenu.Item>
                  <DropdownMenu.Separator />
                  <DropdownMenu.Item
                    class="text-destructive focus:text-destructive"
                    onclick={() => {
                      bookToDelete = book;
                      deleteFile = false;
                    }}
                  >
                    <TrashIcon class="size-3.5" /> Delete
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Portal>
            </DropdownMenu.Root>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<AlertDialog.Root
  open={!!bookToDelete}
  onOpenChange={(o) => {
    if (!o) {
      bookToDelete = null;
      deleteFile = false;
    }
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete "{bookToDelete?.metadata.title}"?</AlertDialog.Title>
      <AlertDialog.Description>
        This will remove the book from your library. This action cannot be undone.
      </AlertDialog.Description>
    </AlertDialog.Header>
    <Label class="flex cursor-pointer items-center gap-2 text-sm">
      <Checkbox bind:checked={deleteFile} />
      Also delete the file from disk
    </Label>
    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
