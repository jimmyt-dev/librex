<script lang="ts">
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { untrack } from 'svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import TagInput from '$lib/components/tag-input.svelte';

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

  const headers = () => ({ Authorization: `Bearer ${getToken()}` });

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
  let editCategories = $state<string[]>([]);
  let editTags = $state<string[]>([]);
  let seriesSuggestions = $state<string[]>([]);
  let showSeriesDropdown = $state(false);
  let seriesHighlightIndex = $state(-1);
  let isSaving = $state(false);
  let errorMsg = $state<string | null>(null);

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
    editCategories = book.categories.map((c) => c.name);
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
        fetch(`/api/books/${bookId}`, { headers: headers() })
          .then((res) => (res.ok ? res.json() : null))
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

  async function fetchAuthorSuggestions(q: string): Promise<string[]> {
    const res = await fetch(`/api/authors?q=${encodeURIComponent(q)}`, { headers: headers() });
    if (!res.ok) return [];
    const data: { name: string }[] = await res.json();
    return data.map((a) => a.name);
  }

  async function fetchCategorySuggestions(q: string): Promise<string[]> {
    const res = await fetch(`/api/categories?q=${encodeURIComponent(q)}`, { headers: headers() });
    if (!res.ok) return [];
    const data: { name: string }[] = await res.json();
    return data.map((c) => c.name);
  }

  async function fetchTagSuggestions(q: string): Promise<string[]> {
    const res = await fetch(`/api/tags?q=${encodeURIComponent(q)}`, { headers: headers() });
    if (!res.ok) return [];
    const data: { name: string }[] = await res.json();
    return data.map((t) => t.name);
  }

  async function fetchSeriesSuggestions(q: string): Promise<string[]> {
    const res = await fetch(`/api/series?q=${encodeURIComponent(q)}`, { headers: headers() });
    if (!res.ok) return [];
    return await res.json();
  }

  async function saveEdit() {
    if (!bookEditState.book) return;
    isSaving = true;
    errorMsg = null;
    try {
      const res = await fetch(`/api/books/${bookEditState.book.id}`, {
        method: 'PUT',
        headers: { ...headers(), 'Content-Type': 'application/json' },
        body: JSON.stringify({
          metadata: {
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
            seriesNumber: editSeriesNumber ? parseFloat(editSeriesNumber) : null
          },
          authors: editAuthors,
          categories: editCategories,
          tags: editTags
        })
      });
      if (!res.ok) throw new Error('Failed to save');
      const updated = await res.json();
      booksState.upsert(updated);
      bookEditState.book = updated;
      bookEditState.close();
    } catch {
      errorMsg = 'Failed to save changes.';
    } finally {
      isSaving = false;
    }
  }
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
            <label for="edit-title" class="text-sm font-medium">Title</label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-subtitle" class="text-sm font-medium">Subtitle</label>
            <Input id="edit-subtitle" bind:value={editSubtitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium">Authors</label>
            <TagInput
              bind:values={editAuthors}
              placeholder="Add author..."
              fetchSuggestions={fetchAuthorSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-description" class="text-sm font-medium">Description</label>
            <Input id="edit-description" bind:value={editDescription} placeholder="Synopsis" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-publisher" class="text-sm font-medium">Publisher</label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <label for="edit-date" class="text-sm font-medium">Published Date</label>
              <Input id="edit-date" bind:value={editPublishedDate} placeholder="YYYY" />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="edit-language" class="text-sm font-medium">Language</label>
              <Input id="edit-language" bind:value={editLanguage} placeholder="en" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <label for="edit-isbn13" class="text-sm font-medium">ISBN-13</label>
              <Input id="edit-isbn13" bind:value={editISBN13} />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="edit-isbn10" class="text-sm font-medium">ISBN-10</label>
              <Input id="edit-isbn10" bind:value={editISBN10} />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-page-count" class="text-sm font-medium">Page Count</label>
            <Input id="edit-page-count" type="number" bind:value={editPageCount} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium">Series</label>
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
                    {#each seriesSuggestions as s, i}
                      <button
                        type="button"
                        class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i === seriesHighlightIndex ? 'bg-accent' : ''}"
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
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium">Categories</label>
            <TagInput
              bind:values={editCategories}
              placeholder="Add category..."
              fetchSuggestions={fetchCategorySuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-medium">Tags</label>
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
          <Button onclick={saveEdit} disabled={isSaving || !editTitle.trim()}>
            {isSaving ? 'Saving…' : 'Save'}
          </Button>
        </Sheet.Footer>
      {/if}
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
