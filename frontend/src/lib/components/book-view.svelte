<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { viewSettings } from '$lib/state/view-settings.svelte';
  import { filterState } from '$lib/state/filter.svelte';
  import { type Book } from '$lib/api/books.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import BookCardSkeleton from '$lib/components/book-card-skeleton.svelte';
  import BookTable from '$lib/components/book-table.svelte';
  import BookViewControls from '$lib/components/book-view-controls.svelte';
  import BookFilterSidebar from '$lib/components/book-filter-sidebar.svelte';
  import SelectionToolbar from '$lib/components/selection-toolbar.svelte';
  import SeriesGroup from '$lib/components/series-group.svelte';
  import { SvelteMap, SvelteSet } from 'svelte/reactivity';
  import { fade } from 'svelte/transition';

  let {
    books,
    isLoading,
    errorMsg = null,
    emptyMessage = 'No books yet.'
  }: {
    books: Book[];
    isLoading: boolean;
    errorMsg?: string | null;
    emptyMessage?: string;
  } = $props();

  let searchQuery = $state('');

  let searchedBooks = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return books;
    return books.filter((b) => {
      const title = (b.metadata.title ?? '').toLowerCase();
      const subtitle = (b.metadata.subtitle ?? '').toLowerCase();
      const series = (b.metadata.seriesName ?? '').toLowerCase();
      const authors = b.authors.map((a) => a.name.toLowerCase()).join(' ');
      return title.includes(q) || subtitle.includes(q) || series.includes(q) || authors.includes(q);
    });
  });

  let sortedBooks = $derived(viewSettings.sort(filterState.apply(searchedBooks)));
  let selectedIds = $state<Set<string>>(new Set());
  let lastSelectedId = $state<string | null>(null);

  let gridClasses = $derived(
    filterState.open
      ? 'grid grid-cols-2 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6'
      : 'grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8'
  );

  let seriesGroups = $derived.by(() => {
    if (!viewSettings.groupBySeries || viewSettings.mode !== 'grid') return null;
    const seriesMap = new SvelteMap<string, Book[]>();
    const ungrouped: Book[] = [];
    for (const book of sortedBooks) {
      const name = book.metadata.seriesName;
      if (name) {
        const existing = seriesMap.get(name);
        if (existing) {
          existing.push(book);
        } else {
          seriesMap.set(name, [book]);
        }
      } else {
        ungrouped.push(book);
      }
    }
    return { seriesMap, ungrouped };
  });

  $effect(() => {
    const bookIds = new Set(books.map((b) => b.id));
    const pruned = new SvelteSet([...selectedIds].filter((id) => bookIds.has(id)));
    if (pruned.size !== selectedIds.size) selectedIds = pruned;
  });

  $effect(() => {
    headerState.counts = isLoading ? [] : [{ label: 'books', value: sortedBooks.length }];
  });

  function toggleSelect(id: string, sel: boolean, shiftKey: boolean) {
    if (shiftKey && sel && lastSelectedId) {
      const ids = sortedBooks.map((b) => b.id);
      const from = ids.indexOf(lastSelectedId);
      const to = ids.indexOf(id);
      const [lo, hi] = from < to ? [from, to] : [to, from];
      const next = new SvelteSet(selectedIds);
      for (let i = lo; i <= hi; i++) next.add(ids[i]);
      selectedIds = next;
    } else {
      const next = new SvelteSet(selectedIds);
      if (sel) next.add(id);
      else next.delete(id);
      selectedIds = next;
    }
    if (sel) lastSelectedId = id;
  }
</script>

<div class="flex flex-1 gap-4 p-4 pt-0">
  <div class="flex min-w-0 flex-1 flex-col gap-4">
    {#if errorMsg}
      <div class="rounded-xl bg-destructive/15 p-4 text-destructive" in:fade>{errorMsg}</div>
    {/if}

    {#if isLoading}
      <div class="flex flex-col gap-4">
        <div
          class="mt-4 flex h-12.5 w-full items-center justify-between gap-2 rounded-md border bg-muted/20 p-2"
        ></div>
        <div class={gridClasses}>
          <!-- eslint-disable-next-line @typescript-eslint/no-unused-vars -->
          {#each Array(12) as _, i (i)}
            <BookCardSkeleton />
          {/each}
        </div>
      </div>
    {:else if books.length === 0}
      <div
        class="mt-4 flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
        in:fade
      >
        <p class="text-muted-foreground">{emptyMessage}</p>
      </div>
    {:else}
      <div class="flex flex-col gap-4" in:fade>
        <BookViewControls bind:searchQuery />
        {#if sortedBooks.length === 0}
          <div
            class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
          >
            <p class="text-muted-foreground">No books match the current filters.</p>
          </div>
        {:else if seriesGroups}
          <!-- Grouped grid view -->
          <div class="flex flex-col gap-6">
            {#each [...seriesGroups.seriesMap.entries()] as [name, seriesBooks] (name)}
              <SeriesGroup
                seriesName={name}
                books={seriesBooks}
                {gridClasses}
                {selectedIds}
                selectMode={selectedIds.size > 0}
                onselect={toggleSelect}
              />
            {/each}
            {#if seriesGroups.ungrouped.length > 0}
              <div class={gridClasses}>
                {#each seriesGroups.ungrouped as book (book.id)}
                  <BookCard
                    {book}
                    selected={selectedIds.has(book.id)}
                    selectMode={selectedIds.size > 0}
                    onselect={toggleSelect}
                  />
                {/each}
              </div>
            {/if}
          </div>
        {:else if viewSettings.mode === 'grid'}
          <div class={gridClasses}>
            {#each sortedBooks as book (book.id)}
              <BookCard
                {book}
                selected={selectedIds.has(book.id)}
                selectMode={selectedIds.size > 0}
                onselect={toggleSelect}
              />
            {/each}
          </div>
        {:else}
          <BookTable
            books={sortedBooks}
            {selectedIds}
            selectMode={selectedIds.size > 0}
            onselect={toggleSelect}
            loading={isLoading}
          />
        {/if}
      </div>
    {/if}
  </div>

  {#if filterState.open}
    <BookFilterSidebar {books} />
  {/if}
</div>

<SelectionToolbar
  {selectedIds}
  books={sortedBooks}
  ondeselect={() => (selectedIds = new Set())}
  onselectall={() => (selectedIds = new Set(sortedBooks.map((b) => b.id)))}
/>
