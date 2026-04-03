<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import {
    fetchAuthorSuggestions,
    fetchGenreSuggestions,
    fetchTagSuggestions,
    fetchSeriesSuggestions
  } from '$lib/api/suggestions';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { untrack } from 'svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import TagInput from '$lib/components/tag-input.svelte';
  import StarRating from '$lib/components/star-rating.svelte';
  import { toast } from 'svelte-sonner';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from './ui/textarea';

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
  let seriesSuggestions = $state<string[]>([]);
  let showSeriesDropdown = $state(false);
  let seriesHighlightIndex = $state(-1);
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
    seriesSuggestions = [];
    showSeriesDropdown = false;
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
        <div
          class="flex flex-col gap-4 overflow-y-auto px-4 py-6"
          oninput={() => (userHasEdited = true)}
        >
          {#if bookEditState.book.metadata.coverPath}
            <img
              src={`/api/books/${bookEditState.book.id}/cover`}
              alt="Cover"
              class="mx-auto h-48 w-auto rounded object-contain shadow"
            />
          {/if}
          {#if errorMsg}
            <p class="text-sm text-destructive">{errorMsg}</p>
          {/if}
          <div class="flex flex-col gap-1.5">
            <Label for="edit-title">
              Title{#if dirtyFields.title}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-subtitle">
              Subtitle{#if dirtyFields.subtitle}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <Input id="edit-subtitle" bind:value={editSubtitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label>
              Authors{#if dirtyFields.authors}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <TagInput
              bind:values={editAuthors}
              placeholder="Add author..."
              fetchSuggestions={fetchAuthorSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-description">
              Description{#if dirtyFields.description}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <Textarea
              id="edit-description"
              bind:value={editDescription}
              placeholder="Description"
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-publisher">
              Publisher{#if dirtyFields.publisher}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <Label for="edit-date">
                Published Date{#if dirtyFields.publishedDate}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </Label>
              <Input id="edit-date" type="date" bind:value={editPublishedDate} />
            </div>
            <div class="flex flex-col gap-1.5">
              <Label for="edit-language">
                Language{#if dirtyFields.language}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </Label>
              <Input id="edit-language" bind:value={editLanguage} placeholder="en" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <Label for="edit-isbn13">
                ISBN-13{#if dirtyFields.isbn13}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </Label>
              <Input id="edit-isbn13" bind:value={editISBN13} />
            </div>
            <div class="flex flex-col gap-1.5">
              <Label for="edit-isbn10">
                ISBN-10{#if dirtyFields.isbn10}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </Label>
              <Input id="edit-isbn10" bind:value={editISBN10} />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-page-count">
              Page Count{#if dirtyFields.pageCount}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <Input id="edit-page-count" type="number" bind:value={editPageCount} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label>
              Series{#if dirtyFields.series}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <div class="grid grid-cols-[1fr_4rem] gap-2">
              <div class="relative">
                <Input
                  bind:value={editSeriesName}
                  placeholder="Series name"
                  oninput={async () => {
                    seriesHighlightIndex = -1;
                    if (editSeriesName.trim().length < 1) {
                      seriesSuggestions = [];
                      showSeriesDropdown = false;
                      return;
                    }
                    seriesSuggestions = await fetchSeriesSuggestions(editSeriesName.trim());
                    showSeriesDropdown = seriesSuggestions.length > 0;
                  }}
                  onkeydown={(e) => {
                    if (e.key === 'ArrowDown') {
                      e.preventDefault();
                      if (seriesSuggestions.length > 0) {
                        showSeriesDropdown = true;
                        seriesHighlightIndex = Math.min(
                          seriesHighlightIndex + 1,
                          seriesSuggestions.length - 1
                        );
                      }
                    } else if (e.key === 'ArrowUp') {
                      e.preventDefault();
                      seriesHighlightIndex = Math.max(seriesHighlightIndex - 1, -1);
                    } else if (e.key === 'Enter' && seriesHighlightIndex >= 0) {
                      e.preventDefault();
                      editSeriesName = seriesSuggestions[seriesHighlightIndex];
                      showSeriesDropdown = false;
                      seriesHighlightIndex = -1;
                    } else if (e.key === 'Escape') {
                      showSeriesDropdown = false;
                      seriesHighlightIndex = -1;
                    }
                  }}
                  onfocus={() => {
                    if (seriesSuggestions.length > 0) showSeriesDropdown = true;
                  }}
                  onblur={() => setTimeout(() => (showSeriesDropdown = false), 150)}
                />
                {#if showSeriesDropdown}
                  <div
                    class="absolute top-full right-0 left-0 z-50 mt-1 max-h-32 overflow-y-auto rounded-lg border bg-popover shadow-md"
                  >
                    {#each seriesSuggestions as s, i (i)}
                      <button
                        type="button"
                        class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i ===
                        seriesHighlightIndex
                          ? 'bg-accent'
                          : ''}"
                        onmousedown={() => {
                          editSeriesName = s;
                          showSeriesDropdown = false;
                          seriesHighlightIndex = -1;
                        }}
                      >
                        {s}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>
              <Input type="number" bind:value={editSeriesNumber} placeholder="#" />
            </div>
            <div class="mt-1 grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <Label class="text-xs font-normal text-muted-foreground">Total Books</Label>
                <Input type="number" bind:value={editSeriesTotal} placeholder="Total" />
              </div>
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <Label>
              Rating{#if dirtyFields.rating}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <StarRating bind:value={editRating} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label>
              Genres{#if dirtyFields.genres}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <TagInput
              bind:values={editGenres}
              placeholder="Add genre..."
              fetchSuggestions={fetchGenreSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label>
              Tags{#if dirtyFields.tags}<span class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </Label>
            <TagInput
              bind:values={editTags}
              placeholder="Add tag..."
              fetchSuggestions={fetchTagSuggestions}
            />
          </div>
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
