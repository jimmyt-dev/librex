<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { untrack } from 'svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { toast } from 'svelte-sonner';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import BookMetaForm from '$lib/components/book-meta-form.svelte';

  let editTitle = $state('');
  let editSubtitle = $state('');
  let editAuthors = $state<string[]>([]);
  let editDescription = $state('');
  let editPublisher = $state('');
  let editPublishedDate = $state('');
  let editISBN13 = $state('');
  let editISBN10 = $state('');
  let editLanguage = $state('');
  let editPageCount = $state('');
  let editSeriesName = $state('');
  let editSeriesNumber = $state('');
  let editSeriesTotal = $state('');
  let editRating = $state('');
  let editGenres = $state<string[]>([]);
  let editTags = $state<string[]>([]);

  function normalizeDate(d: string | null | undefined): string {
    if (!d) return '';
    if (/^\d{4}$/.test(d)) return `${d}-01-01`;
    if (/^\d{4}-\d{2}$/.test(d)) return `${d}-01`;
    return d;
  }

  let isSaving = $state(false);
  let errorMsg = $state<string | null>(null);
  let userHasEdited = $state(false);

  let dirtyFields = $derived.by(() => {
    const b = bookEditState.book;
    if (!b) return {} as Record<string, boolean>;
    return {
      title: editTitle !== b.metadata.title,
      subtitle: editSubtitle !== (b.metadata.subtitle ?? ''),
      authors: JSON.stringify(editAuthors) !== JSON.stringify(b.authors.map((a) => a.name)),
      description: editDescription !== (b.metadata.description ?? ''),
      publisher: editPublisher !== (b.metadata.publisher ?? ''),
      publishedDate: editPublishedDate !== normalizeDate(b.metadata.publishedDate),
      isbn13: editISBN13 !== (b.metadata.isbn13 ?? ''),
      isbn10: editISBN10 !== (b.metadata.isbn10 ?? ''),
      language: editLanguage !== (b.metadata.language ?? ''),
      pageCount: editPageCount !== (b.metadata.pageCount?.toString() ?? ''),
      series:
        editSeriesName !== (b.metadata.seriesName ?? '') ||
        editSeriesNumber !== (b.metadata.seriesNumber?.toString() ?? '') ||
        editSeriesTotal !== (b.metadata.seriesTotal?.toString() ?? ''),
      rating: editRating !== (b.metadata.rating?.toString() ?? ''),
      genres: JSON.stringify(editGenres) !== JSON.stringify(b.genres.map((g) => g.name)),
      tags: JSON.stringify(editTags) !== JSON.stringify(b.tags.map((t) => t.name))
    };
  });

  let dirty = $derived(Object.values(dirtyFields).some(Boolean));

  function resetToBook() {
    const book = bookEditState.book;
    if (!book) return;
    editTitle = book.metadata.title;
    editSubtitle = book.metadata.subtitle ?? '';
    editAuthors = book.authors.map((a) => a.name);
    editDescription = book.metadata.description ?? '';
    editPublisher = book.metadata.publisher ?? '';
    editPublishedDate = normalizeDate(book.metadata.publishedDate);
    editISBN13 = book.metadata.isbn13 ?? '';
    editISBN10 = book.metadata.isbn10 ?? '';
    editLanguage = book.metadata.language ?? '';
    editPageCount = book.metadata.pageCount?.toString() ?? '';
    editSeriesName = book.metadata.seriesName ?? '';
    editSeriesNumber = book.metadata.seriesNumber?.toString() ?? '';
    editSeriesTotal = book.metadata.seriesTotal?.toString() ?? '';
    editRating = book.metadata.rating?.toString() ?? '';
    editGenres = book.genres.map((g) => g.name);
    editTags = book.tags.map((t) => t.name);
    errorMsg = null;
    userHasEdited = false;
  }

  // Derive a primitive string so the $effect tracks queue navigation without
  // triggering an infinite loop when bookEditState.book = fresh (same id, new ref).
  let bookId = $derived(bookEditState.book?.id ?? '');

  $effect(() => {
    if (!bookEditState.open || !bookId) return;
    untrack(() => {
      resetToBook();
      apiFetch(`/api/books/${bookId}`)
        .then((fresh) => {
          if (fresh && bookEditState.open && bookEditState.book?.id === bookId) {
            booksState.upsert(fresh);
            bookEditState.book = fresh;
            // Only reset if the user hasn't intentionally edited anything yet.
            // `userHasEdited` is cleared by resetToBook() and set by any oninput in the form.
            if (!userHasEdited) resetToBook();
          }
        })
        .catch(() => {});
    });
  });

  // Refresh all books when the sheet closes so the list stays in sync.
  let _wasOpen = false;
  $effect(() => {
    const isOpen = bookEditState.open;
    if (!isOpen && _wasOpen) {
      untrack(() => booksState.fetchAll());
    }
    _wasOpen = isOpen;
  });

  function handleKeyDown(event: KeyboardEvent) {
    if (!bookEditState.inQueue) return;
    if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement)
      return;
    if (event.key === 'ArrowRight') {
      event.preventDefault();
      if (bookEditState.isLast) {
        bookEditState.goTo(0);
      } else {
        bookEditState.next();
      }
    } else if (event.key === 'ArrowLeft') {
      event.preventDefault();
      if (bookEditState.isFirst) {
        bookEditState.goTo(bookEditState.queue.length - 1);
      } else {
        bookEditState.prev();
      }
    }
  }

  async function saveEdit() {
    if (!bookEditState.book) return;
    const book = bookEditState.book;
    const originalBook = JSON.parse(JSON.stringify(book));

    // Construct updated book
    const updated: Book = {
      ...book,
      metadata: {
        ...book.metadata,
        title: editTitle.trim(),
        subtitle: editSubtitle.trim(),
        description: editDescription.trim(),
        publisher: editPublisher.trim(),
        publishedDate: editPublishedDate.trim(),
        isbn13: editISBN13.trim(),
        isbn10: editISBN10.trim(),
        language: editLanguage.trim(),
        pageCount: editPageCount ? parseInt(editPageCount) : null,
        seriesName: editSeriesName.trim(),
        seriesNumber: editSeriesNumber ? parseFloat(editSeriesNumber) : null,
        seriesTotal: editSeriesTotal ? parseInt(editSeriesTotal) : null,
        rating: editRating ? parseInt(editRating) : null
      },
      authors: editAuthors.map((name) => ({ id: name, name })),
      genres: editGenres.map((name) => ({ id: name, name })),
      tags: editTags.map((name) => ({ id: name, name }))
    };

    // Optimistic update
    booksState.upsert(updated);

    // Advance through queue (stay open on last book)
    if (bookEditState.inQueue) {
      if (!bookEditState.isLast) bookEditState.next();
    } else {
      bookEditState.close();
    }

    isSaving = true;
    errorMsg = null;
    try {
      const serverUpdated = await apiFetch(`/api/books/${book.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          metadata: updated.metadata,
          authors: editAuthors,
          genres: editGenres,
          tags: editTags
        })
      });
      booksState.upsert(serverUpdated);
      toast.success('Changes saved.');
      // If we're still showing the book that was just saved (last in queue, or
      // single-book edit), update the baseline so dirty/revert clear correctly.
      if (bookEditState.book?.id === book.id) {
        bookEditState.book = serverUpdated;
        resetToBook();
      }
    } catch {
      // Revert
      booksState.upsert(originalBook);
      toast.error('Failed to save changes.');
    } finally {
      isSaving = false;
    }
  }

  export const _isSheet = true;
</script>

<svelte:window onkeydown={handleKeyDown} />

<Sheet.Root bind:open={bookEditState.open}>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      {#if bookEditState.book}
        <Sheet.Header>
          <div class="flex items-center justify-between gap-2">
            <Sheet.Title>Edit Metadata</Sheet.Title>
          </div>
          <Sheet.Description class="truncate text-xs text-muted-foreground">
            {bookEditState.book.filePath.split('/').pop()}
          </Sheet.Description>
        </Sheet.Header>
        <div class="overflow-y-auto px-4 py-6">
          {#if errorMsg}
            <p class="mb-4 text-sm text-destructive">{errorMsg}</p>
          {/if}
          <BookMetaForm
            bind:title={editTitle}
            bind:subtitle={editSubtitle}
            bind:authors={editAuthors}
            bind:description={editDescription}
            bind:publisher={editPublisher}
            bind:publishedDate={editPublishedDate}
            bind:isbn13={editISBN13}
            bind:isbn10={editISBN10}
            bind:language={editLanguage}
            bind:pageCount={editPageCount}
            bind:seriesName={editSeriesName}
            bind:seriesNumber={editSeriesNumber}
            bind:seriesTotal={editSeriesTotal}
            bind:rating={editRating}
            bind:genres={editGenres}
            bind:tags={editTags}
            coverSrc={bookEditState.book.metadata.coverPath
              ? `/api/books/${bookEditState.book.id}/cover`
              : null}
            {dirtyFields}
            oninput={() => (userHasEdited = true)}
          />
        </div>
        <Sheet.Footer>
          {#if bookEditState.inQueue}
            <div class="mr-4 flex items-center gap-0.5 text-muted-foreground">
              <Button
                size="icon"
                variant="outline"
                onclick={() => {
                  if (bookEditState.isFirst) {
                    bookEditState.goTo(bookEditState.queue.length - 1);
                  } else {
                    bookEditState.prev();
                  }
                }}
                title="Previous book"
              >
                <ChevronLeftIcon class="size-4" />
              </Button>
              <span class="min-w-10 text-center text-xs tabular-nums">
                {bookEditState.queueIndex + 1}&nbsp;/&nbsp;{bookEditState.queue.length}
              </span>
              <Button
                size="icon"
                variant="outline"
                onclick={() => {
                  if (bookEditState.isLast) {
                    bookEditState.goTo(0);
                  } else {
                    bookEditState.next();
                  }
                }}
                title="Next book"
              >
                <ChevronRightIcon class="size-4" />
              </Button>
            </div>
          {/if}
          {#if dirty && userHasEdited && !isSaving}
            <Button variant="outline" onclick={resetToBook}>
              <RotateCcwIcon class="size-4" />
              Revert
            </Button>
          {/if}
          <Sheet.Close>
            {#snippet child({ props })}
              <Button variant="outline" {...props} onclick={() => bookEditState.close()}>
                Cancel
              </Button>
            {/snippet}
          </Sheet.Close>
          <Button onclick={saveEdit} disabled={isSaving || !editTitle.trim()}>
            {#if isSaving}
              Saving…
            {:else if bookEditState.inQueue && !bookEditState.isLast}
              Save & Next
            {:else}
              Save
            {/if}
          </Button>
        </Sheet.Footer>
      {/if}
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
