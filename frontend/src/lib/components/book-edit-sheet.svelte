<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import {
    fetchAuthorSuggestions,
    fetchGenreSuggestions,
    fetchTagSuggestions,
    fetchSeriesSuggestions
  } from '$lib/api/suggestions';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { untrack } from 'svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import TagInput from '$lib/components/tag-input.svelte';
  import StarRating from '$lib/components/star-rating.svelte';
  import { toast } from 'svelte-sonner';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';

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
  let isSaving = $state(false);
  let errorMsg = $state<string | null>(null);

  let dirtyFields = $derived.by(() => {
    const b = bookEditState.book;
    if (!b) return {} as Record<string, boolean>;
    return {
      title: editTitle !== b.metadata.title,
      subtitle: editSubtitle !== (b.metadata.subtitle ?? ''),
      authors: JSON.stringify(editAuthors) !== JSON.stringify(b.authors.map((a) => a.name)),
      description: editDescription !== (b.metadata.description ?? ''),
      publisher: editPublisher !== (b.metadata.publisher ?? ''),
      publishedDate: editPublishedDate !== (b.metadata.publishedDate ?? ''),
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
    editPublishedDate = book.metadata.publishedDate ?? '';
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
  }

  $effect(() => {
    if (bookEditState.open) {
      untrack(() => {
        const bookId = bookEditState.book?.id;
        if (!bookId) return;
        // Fetch fresh data from API to avoid stale snapshots
        apiFetch(`/api/books/${bookId}`)
          .then((fresh) => {
            if (fresh && bookEditState.open) {
              bookEditState.book = fresh;
              resetToBook();
            }
          })
          .catch(() => resetToBook());
      });
    }
  });

  async function saveEdit() {
    if (!bookEditState.book) return;
    const book = bookEditState.book;
    const originalBook = JSON.parse(JSON.stringify(book));

    // Construct updated book
    const updated: any = {
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
      authors: editAuthors.map((name) => ({ id: '', name })), // Placeholder IDs
      genres: editGenres.map((name) => ({ id: '', name })),
      tags: editTags.map((name) => ({ id: '', name }))
    };

    // Optimistic update
    booksState.upsert(updated);
    bookEditState.close();

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

<Sheet.Root bind:open={bookEditState.open}>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      {#if bookEditState.book}
        <Sheet.Header>
          <Sheet.Title>Edit Metadata</Sheet.Title>
          <Sheet.Description class="truncate text-xs text-muted-foreground">
            {bookEditState.book.filePath.split('/').pop()}
          </Sheet.Description>
        </Sheet.Header>
        <div class="flex flex-col gap-4 overflow-y-auto px-4 py-6">
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
            <label for="edit-title" class="flex items-center gap-1.5 text-sm font-medium">
              Title{#if dirtyFields.title}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-subtitle" class="flex items-center gap-1.5 text-sm font-medium">
              Subtitle{#if dirtyFields.subtitle}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <Input id="edit-subtitle" bind:value={editSubtitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="flex items-center gap-1.5 text-sm font-medium">
              Authors{#if dirtyFields.authors}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <TagInput
              bind:values={editAuthors}
              placeholder="Add author..."
              fetchSuggestions={fetchAuthorSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-description" class="flex items-center gap-1.5 text-sm font-medium">
              Description{#if dirtyFields.description}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <Input id="edit-description" bind:value={editDescription} placeholder="Synopsis" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-publisher" class="flex items-center gap-1.5 text-sm font-medium">
              Publisher{#if dirtyFields.publisher}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <label for="edit-date" class="flex items-center gap-1.5 text-sm font-medium">
                Published Date{#if dirtyFields.publishedDate}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </label>
              <Input id="edit-date" bind:value={editPublishedDate} placeholder="YYYY" />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="edit-language" class="flex items-center gap-1.5 text-sm font-medium">
                Language{#if dirtyFields.language}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </label>
              <Input id="edit-language" bind:value={editLanguage} placeholder="en" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <label for="edit-isbn13" class="flex items-center gap-1.5 text-sm font-medium">
                ISBN-13{#if dirtyFields.isbn13}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </label>
              <Input id="edit-isbn13" bind:value={editISBN13} />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="edit-isbn10" class="flex items-center gap-1.5 text-sm font-medium">
                ISBN-10{#if dirtyFields.isbn10}<span
                    class="inline-block size-1.5 rounded-full bg-primary"
                  ></span>{/if}
              </label>
              <Input id="edit-isbn10" bind:value={editISBN10} />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-page-count" class="flex items-center gap-1.5 text-sm font-medium">
              Page Count{#if dirtyFields.pageCount}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <Input id="edit-page-count" type="number" bind:value={editPageCount} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="flex items-center gap-1.5 text-sm font-medium">
              Series{#if dirtyFields.series}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
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
                <label class="text-xs text-muted-foreground">Total Books</label>
                <Input type="number" bind:value={editSeriesTotal} placeholder="Total" />
              </div>
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="flex items-center gap-1.5 text-sm font-medium">
              Rating{#if dirtyFields.rating}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <StarRating bind:value={editRating} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="flex items-center gap-1.5 text-sm font-medium">
              Genres{#if dirtyFields.genres}<span
                  class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <TagInput
              bind:values={editGenres}
              placeholder="Add genre..."
              fetchSuggestions={fetchGenreSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="flex items-center gap-1.5 text-sm font-medium">
              Tags{#if dirtyFields.tags}<span class="inline-block size-1.5 rounded-full bg-primary"
                ></span>{/if}
            </label>
            <TagInput
              bind:values={editTags}
              placeholder="Add tag..."
              fetchSuggestions={fetchTagSuggestions}
            />
          </div>
        </div>
        <Sheet.Footer>
          <Sheet.Close>
            {#snippet child({ props })}
              <Button variant="outline" {...props}>Cancel</Button>
            {/snippet}
          </Sheet.Close>
          {#if dirty}
            <Button variant="ghost" onclick={resetToBook}>
              <RotateCcwIcon class="size-4" />
              Revert
            </Button>
          {/if}
          <Button onclick={saveEdit} disabled={isSaving || !editTitle.trim()}>
            {isSaving ? 'Saving…' : 'Save'}
          </Button>
        </Sheet.Footer>
      {/if}
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
