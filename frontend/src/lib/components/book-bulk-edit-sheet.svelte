<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import {
    fetchAuthorSuggestions,
    fetchGenreSuggestions,
    fetchTagSuggestions,
    fetchSeriesSuggestions,
    fetchPublisherSuggestions
  } from '$lib/api/suggestions';
  import { booksState, type Book } from '$lib/api/books.svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import StarRating from '$lib/components/star-rating.svelte';
  import ArrayField from '$lib/components/array-field.svelte';
  import { Label } from './ui/label';

  let {
    open = $bindable(false),
    selectedIds,
    books
  }: {
    open: boolean;
    selectedIds: Set<string>;
    books: Book[];
  } = $props();

  // Text fields — empty string means "don't change"
  let seriesName = $state('');
  let seriesSuggestions = $state<string[]>([]);
  let showSeriesDropdown = $state(false);
  let seriesHighlightIndex = $state(-1);
  let publisher = $state('');
  let publisherSuggestions = $state<string[]>([]);
  let showPublisherDropdown = $state(false);
  let publisherHighlightIndex = $state(-1);
  let language = $state('');
  let seriesTotal = $state('');
  let rating = $state('');

  // Array fields
  let authors = $state<string[]>([]);
  let authorsMode = $state<'replace' | 'merge'>('merge');
  let genres = $state<string[]>([]);
  let genresMode = $state<'replace' | 'merge'>('merge');
  let tags = $state<string[]>([]);
  let tagsMode = $state<'replace' | 'merge'>('merge');

  let saving = $state(false);
  let errorMsg = $state<string | null>(null);

  function reset() {
    seriesName = '';
    publisher = '';
    language = '';
    seriesTotal = '';
    rating = '';
    authors = [];
    authorsMode = 'merge';
    genres = [];
    genresMode = 'merge';
    tags = [];
    tagsMode = 'merge';
    errorMsg = null;
  }

  $effect(() => {
    if (!open) reset();
  });

  function hasChanges() {
    return (
      seriesName.trim() !== '' ||
      publisher.trim() !== '' ||
      language.trim() !== '' ||
      seriesTotal !== '' ||
      rating !== '' ||
      authors.length > 0 ||
      genres.length > 0 ||
      tags.length > 0
    );
  }

  async function save() {
    if (!hasChanges()) return;
    saving = true;
    errorMsg = null;
    try {
      const payload: Record<string, unknown> = {
        bookIds: [...selectedIds]
      };
      if (seriesName.trim() !== '') payload.seriesName = seriesName.trim();
      if (publisher.trim() !== '') payload.publisher = publisher.trim();
      if (language.trim() !== '') payload.language = language.trim();
      if (seriesTotal !== '') payload.seriesTotal = parseInt(seriesTotal);
      if (rating !== '') payload.rating = parseInt(rating);
      if (authors.length > 0) {
        payload.authors = authors;
        payload.authorsMode = authorsMode;
      }
      if (genres.length > 0) {
        payload.genres = genres;
        payload.genresMode = genresMode;
      }
      if (tags.length > 0) {
        payload.tags = tags;
        payload.tagsMode = tagsMode;
      }

      await apiFetch('/api/books/bulk-update', {
        method: 'POST',
        body: JSON.stringify(payload)
      });

      // Refresh affected books
      const affectedLibraryIds = new Set(
        books.filter((b) => selectedIds.has(b.id)).map((b) => b.libraryId)
      );
      await booksState.fetchAll();
      for (const libId of affectedLibraryIds) {
        booksState.fetchForLibrary(libId);
      }

      open = false;
    } catch {
      errorMsg = 'Failed to apply changes.';
    } finally {
      saving = false;
    }
  }
</script>

<Sheet.Root bind:open>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      <Sheet.Header>
        <Sheet.Title>Bulk Edit</Sheet.Title>
        <Sheet.Description class="text-xs text-muted-foreground">
          Editing {selectedIds.size} book{selectedIds.size > 1 ? 's' : ''}. Leave a field empty to
          leave it unchanged.
        </Sheet.Description>
      </Sheet.Header>

      <div class="flex flex-col gap-4 overflow-y-auto px-4 py-6">
        {#if errorMsg}
          <p class="text-sm text-destructive">{errorMsg}</p>
        {/if}

        <!-- Text fields -->
        <div class="relative flex flex-col gap-1.5">
          <Label for="bulk-series" class="text-sm font-medium">Series Name</Label>
          <Input
            id="bulk-series"
            bind:value={seriesName}
            placeholder="Series name"
            oninput={async () => {
              seriesHighlightIndex = -1;
              if (seriesName.trim().length < 1) {
                seriesSuggestions = [];
                showSeriesDropdown = false;
                return;
              }
              seriesSuggestions = await fetchSeriesSuggestions(seriesName.trim());
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
                seriesName = seriesSuggestions[seriesHighlightIndex];
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
                    seriesName = s;
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

        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-1.5">
            <Label for="bulk-series-total" class="text-sm font-medium">Series Total</Label>
            <Input id="bulk-series-total" type="number" bind:value={seriesTotal} placeholder="—" />
          </div>
          <!-- <div class="flex flex-col gap-1.5">
            <Label for="bulk-publisher" class="text-sm font-medium">Publisher</Label>
            <Input id="bulk-publisher" bind:value={publisher} placeholder="Leave empty to skip" />
          </div> -->
          <div class="relative flex flex-col gap-1.5">
            <Label for="bulk-publisher" class="text-sm font-medium">Publisher</Label>
            <Input
              id="bulk-publisher"
              bind:value={publisher}
              placeholder="Publisher name"
              oninput={async () => {
                publisherHighlightIndex = -1;
                if (publisher.trim().length < 1) {
                  publisherSuggestions = [];
                  showPublisherDropdown = false;
                  return;
                }
                publisherSuggestions = await fetchPublisherSuggestions(publisher.trim());
                showPublisherDropdown = publisherSuggestions.length > 0;
              }}
              onkeydown={(e) => {
                if (e.key === 'ArrowDown') {
                  e.preventDefault();
                  if (publisherSuggestions.length > 0) {
                    showPublisherDropdown = true;
                    publisherHighlightIndex = Math.min(
                      publisherHighlightIndex + 1,
                      publisherSuggestions.length - 1
                    );
                  }
                } else if (e.key === 'ArrowUp') {
                  e.preventDefault();
                  publisherHighlightIndex = Math.max(publisherHighlightIndex - 1, -1);
                } else if (e.key === 'Enter' && publisherHighlightIndex >= 0) {
                  e.preventDefault();
                  publisher = publisherSuggestions[publisherHighlightIndex];
                  showPublisherDropdown = false;
                  publisherHighlightIndex = -1;
                } else if (e.key === 'Escape') {
                  showPublisherDropdown = false;
                  publisherHighlightIndex = -1;
                }
              }}
              onfocus={() => {
                if (publisherSuggestions.length > 0) showPublisherDropdown = true;
              }}
              onblur={() => setTimeout(() => (showPublisherDropdown = false), 150)}
            />
            {#if showPublisherDropdown}
              <div
                class="absolute top-full right-0 left-0 z-50 mt-1 max-h-32 overflow-y-auto rounded-lg border bg-popover shadow-md"
              >
                {#each publisherSuggestions as s, i (i)}
                  <button
                    type="button"
                    class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i ===
                    publisherHighlightIndex
                      ? 'bg-accent'
                      : ''}"
                    onmousedown={() => {
                      publisher = s;
                      showPublisherDropdown = false;
                      publisherHighlightIndex = -1;
                    }}
                  >
                    {s}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <Label for="bulk-language" class="text-sm font-medium">Language</Label>
          <Input id="bulk-language" bind:value={language} placeholder="e.g. en" />
        </div>

        <!-- Rating -->
        <div class="flex flex-col gap-1.5">
          <Label class="text-sm font-medium">Rating</Label>
          <StarRating bind:value={rating} />
        </div>

        <div class="h-px bg-border"></div>

        <!-- Array fields -->
        <ArrayField
          label="Authors"
          bind:values={authors}
          bind:mode={authorsMode}
          placeholder="Add author..."
          fetchSuggestions={fetchAuthorSuggestions}
        />
        <ArrayField
          label="Genres"
          bind:values={genres}
          bind:mode={genresMode}
          placeholder="Add genre..."
          fetchSuggestions={fetchGenreSuggestions}
        />
        <ArrayField
          label="Tags"
          bind:values={tags}
          bind:mode={tagsMode}
          placeholder="Add tag..."
          fetchSuggestions={fetchTagSuggestions}
        />
      </div>

      <Sheet.Footer>
        <Sheet.Close>
          {#snippet child({ props })}
            <Button variant="outline" {...props}>Cancel</Button>
          {/snippet}
        </Sheet.Close>
        <Button onclick={save} disabled={saving || !hasChanges()}>
          {saving
            ? 'Saving…'
            : `Apply to ${selectedIds.size} book${selectedIds.size > 1 ? 's' : ''}`}
        </Button>
      </Sheet.Footer>
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
