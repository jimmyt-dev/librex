<script lang="ts">
  import type { Book } from '$lib/api/books.svelte';
  import { filterState, type FilterMode, type ItemState } from '$lib/state/filter.svelte';
  import { Button } from '$lib/components/ui/button';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
  import CheckIcon from '@lucide/svelte/icons/check';
  import StarIcon from '@lucide/svelte/icons/star';
  import XIcon from '@lucide/svelte/icons/x';
  import { SvelteSet } from 'svelte/reactivity';

  let { books }: { books: Book[] } = $props();

  let availableAuthors = $derived.by(() => {
    const names = new SvelteSet<string>();
    for (const b of books) for (const a of b.authors) names.add(a.name);
    return [...names].sort((a, b) => a.localeCompare(b));
  });

  let availableGenres = $derived.by(() => {
    const names = new SvelteSet<string>();
    for (const b of books) for (const g of b.genres) names.add(g.name);
    return [...names].sort((a, b) => a.localeCompare(b));
  });

  let availableTags = $derived.by(() => {
    const names = new SvelteSet<string>();
    for (const b of books) for (const t of b.tags) names.add(t.name);
    return [...names].sort((a, b) => a.localeCompare(b));
  });

  let availableLanguages = $derived.by(() => {
    const langs = new SvelteSet<string>();
    for (const b of books) if (b.metadata.language) langs.add(b.metadata.language);
    return [...langs].sort();
  });

  const STATUSES = [
    { value: 'unread', label: 'Not Started' },
    { value: 'reading', label: 'Reading' },
    { value: 'finished', label: 'Finished' },
    { value: 'dnf', label: 'Did Not Finish' }
  ];

  let sectionOpen = $state({
    status: true,
    rating: true,
    authors: true,
    genres: true,
    tags: true,
    language: true
  });

  function langLabel(code: string): string {
    try {
      return new Intl.DisplayNames(['en'], { type: 'language' }).of(code) ?? code;
    } catch {
      return code;
    }
  }
</script>

{#snippet triItem(label: string, state: ItemState | undefined, ontoggle: () => void)}
  <button
    type="button"
    class="flex w-full items-center gap-2 rounded px-1 py-0.5 text-left hover:bg-muted/50"
    onclick={ontoggle}
    title={state === 'include'
      ? 'Click to exclude'
      : state === 'exclude'
        ? 'Click to clear'
        : 'Click to include'}
  >
    {#if state === 'include'}
      <div
        class="flex size-4 shrink-0 items-center justify-center rounded border border-primary bg-primary"
      >
        <CheckIcon class="size-2.5 text-primary-foreground" />
      </div>
    {:else if state === 'exclude'}
      <div
        class="flex size-4 shrink-0 items-center justify-center rounded border border-destructive bg-destructive"
      >
        <XIcon class="text-destructive-foreground size-2.5" />
      </div>
    {:else}
      <div class="size-4 shrink-0 rounded border border-input"></div>
    {/if}
    <span
      class="truncate text-xs {state === 'exclude' ? 'text-muted-foreground line-through' : ''}"
    >
      {label}
    </span>
  </button>
{/snippet}

{#snippet modeToggle(mode: FilterMode, onchange: (m: FilterMode) => void)}
  <div class="flex items-center justify-between px-1 pt-0.5 pb-1.5">
    <span class="text-xs text-muted-foreground">Match</span>
    <div class="flex overflow-hidden rounded border text-xs">
      <button
        type="button"
        class="px-2 py-0.5 transition-colors {mode === 'or'
          ? 'bg-primary text-primary-foreground'
          : 'text-muted-foreground hover:bg-muted'}"
        onclick={() => onchange('or')}
      >
        OR
      </button>
      <button
        type="button"
        class="px-2 py-0.5 transition-colors {mode === 'and'
          ? 'bg-primary text-primary-foreground'
          : 'text-muted-foreground hover:bg-muted'}"
        onclick={() => onchange('and')}
      >
        AND
      </button>
    </div>
  </div>
{/snippet}

{#snippet sectionHeader(title: string, count: number, open: boolean, ontoggle: () => void)}
  <button
    type="button"
    class="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-sm font-medium hover:bg-muted/50"
    onclick={ontoggle}
  >
    {title}
    {#if count > 0}
      <span class="rounded-full bg-primary px-1.5 text-xs text-primary-foreground">{count}</span>
    {:else}
      <ChevronDownIcon
        class="size-3.5 text-muted-foreground transition-transform {open ? '' : '-rotate-90'}"
      />
    {/if}
  </button>
{/snippet}

<aside class="sticky top-4 flex w-64 shrink-0 flex-col self-start rounded-lg border bg-card">
  <div class="flex items-center justify-between border-b px-4 py-3">
    <span class="text-sm font-semibold">Filters</span>
    {#if filterState.activeCount > 0}
      <Button
        variant="ghost"
        size="sm"
        class="h-6 px-2 text-xs"
        onclick={() => filterState.clear()}
      >
        Clear all
      </Button>
    {/if}
  </div>

  <div class="flex flex-col overflow-y-auto p-2">
    <!-- Reading Status -->
    <div>
      {@render sectionHeader(
        'Status',
        filterState.statusSelections.size,
        sectionOpen.status,
        () => (sectionOpen.status = !sectionOpen.status)
      )}
      {#if sectionOpen.status}
        <div class="mt-0.5 flex flex-col pb-2">
          {#each STATUSES as s (s.value)}
            {@render triItem(s.label, filterState.statusSelections.get(s.value), () =>
              filterState.toggleStatus(s.value)
            )}
          {/each}
        </div>
      {/if}
    </div>

    <div class="my-1 border-t"></div>

    <!-- Rating -->
    <div>
      <button
        type="button"
        class="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-sm font-medium hover:bg-muted/50"
        onclick={() => (sectionOpen.rating = !sectionOpen.rating)}
      >
        Min. Rating
        {#if filterState.minRating !== null}
          <span class="rounded-full bg-primary px-1.5 text-xs text-primary-foreground">
            {filterState.minRating}+
          </span>
        {:else}
          <ChevronDownIcon
            class="size-3.5 text-muted-foreground transition-transform {sectionOpen.rating
              ? ''
              : '-rotate-90'}"
          />
        {/if}
      </button>
      {#if sectionOpen.rating}
        <div class="flex items-center gap-0.5 px-2 pb-2">
          {#each [1, 2, 3, 4, 5] as star (star)}
            <button
              type="button"
              class="rounded p-1 transition-colors hover:bg-muted"
              onclick={() => filterState.setRating(star)}
              title="{star} star minimum"
            >
              <StarIcon
                class="size-4 {filterState.minRating !== null && star <= filterState.minRating
                  ? 'fill-yellow-400 text-yellow-400'
                  : 'text-muted-foreground'}"
              />
            </button>
          {/each}
          {#if filterState.minRating !== null}
            <button
              type="button"
              class="ml-auto rounded p-1 text-muted-foreground hover:text-foreground"
              onclick={() => (filterState.minRating = null)}
              title="Clear rating filter"
            >
              <XIcon class="size-3.5" />
            </button>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Authors -->
    {#if availableAuthors.length > 0}
      <div class="my-1 border-t"></div>
      <div>
        {@render sectionHeader(
          'Authors',
          filterState.authorSelections.size,
          sectionOpen.authors,
          () => (sectionOpen.authors = !sectionOpen.authors)
        )}
        {#if sectionOpen.authors}
          {@render modeToggle(filterState.authorMode, (m) => (filterState.authorMode = m))}
          <div class="max-h-48 overflow-y-auto pb-2">
            <div class="flex flex-col">
              {#each availableAuthors as name (name)}
                {@render triItem(name, filterState.authorSelections.get(name), () =>
                  filterState.toggleAuthor(name)
                )}
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Genres -->
    {#if availableGenres.length > 0}
      <div class="my-1 border-t"></div>
      <div>
        {@render sectionHeader(
          'Genres',
          filterState.genreSelections.size,
          sectionOpen.genres,
          () => (sectionOpen.genres = !sectionOpen.genres)
        )}
        {#if sectionOpen.genres}
          {@render modeToggle(filterState.genreMode, (m) => (filterState.genreMode = m))}
          <div class="max-h-48 overflow-y-auto pb-2">
            <div class="flex flex-col">
              {#each availableGenres as name (name)}
                {@render triItem(name, filterState.genreSelections.get(name), () =>
                  filterState.toggleGenre(name)
                )}
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Tags -->
    {#if availableTags.length > 0}
      <div class="my-1 border-t"></div>
      <div>
        {@render sectionHeader(
          'Tags',
          filterState.tagSelections.size,
          sectionOpen.tags,
          () => (sectionOpen.tags = !sectionOpen.tags)
        )}
        {#if sectionOpen.tags}
          {@render modeToggle(filterState.tagMode, (m) => (filterState.tagMode = m))}
          <div class="max-h-48 overflow-y-auto pb-2">
            <div class="flex flex-col">
              {#each availableTags as name (name)}
                {@render triItem(name, filterState.tagSelections.get(name), () =>
                  filterState.toggleTag(name)
                )}
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Language -->
    {#if availableLanguages.length > 0}
      <div class="my-1 border-t"></div>
      <div>
        {@render sectionHeader(
          'Language',
          filterState.languageSelections.size,
          sectionOpen.language,
          () => (sectionOpen.language = !sectionOpen.language)
        )}
        {#if sectionOpen.language}
          <div class="mt-0.5 flex flex-col pb-2">
            {#each availableLanguages as lang (lang)}
              {@render triItem(langLabel(lang), filterState.languageSelections.get(lang), () =>
                filterState.toggleLanguage(lang)
              )}
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</aside>
