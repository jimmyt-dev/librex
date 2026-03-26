<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import { viewSettings } from '$lib/state/view-settings.svelte';
  import { filterState } from '$lib/state/filter.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import BookTable from '$lib/components/book-table.svelte';
  import BookViewControls from '$lib/components/book-view-controls.svelte';
  import BookFilterSidebar from '$lib/components/book-filter-sidebar.svelte';
  import SelectionToolbar from '$lib/components/selection-toolbar.svelte';
  import { SvelteSet } from 'svelte/reactivity';

  let sortedBooks = $derived(viewSettings.sort(filterState.apply(booksState.all)));

  headerState.title = 'All Books';
  headerState.subtitle = null;

  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);
  let selectedIds = $state<Set<string>>(new Set());
  let lastSelectedId = $state<string | null>(null);

  $effect(() => {
    isLoading = true;
    errorMsg = null;
    booksState
      .fetchAll()
      .catch((err: unknown) => {
        errorMsg = err instanceof Error ? err.message : 'Failed to load books.';
      })
      .finally(() => {
        isLoading = false;
      });
  });

  $effect(() => {
    headerState.counts = isLoading ? [] : [{ label: 'books', value: sortedBooks.length }];
  });

  function toggleSelect(id: string, selected: boolean, shiftKey: boolean) {
    const books = booksState.all;
    if (shiftKey && selected && lastSelectedId) {
      const ids = books.map((b) => b.id);
      const from = ids.indexOf(lastSelectedId);
      const to = ids.indexOf(id);
      const [lo, hi] = from < to ? [from, to] : [to, from];
      const next = new SvelteSet(selectedIds);
      for (let i = lo; i <= hi; i++) next.add(ids[i]);
      selectedIds = next;
    } else {
      const next = new SvelteSet(selectedIds);
      if (selected) next.add(id);
      else next.delete(id);
      selectedIds = next;
    }
    if (selected) lastSelectedId = id;
  }
</script>

<div class="flex flex-1 gap-4 p-4 pt-0">
  <div class="flex min-w-0 flex-1 flex-col gap-4">
    {#if isLoading}
      <div class="flex min-h-64 items-center justify-center">
        <p class="text-muted-foreground">Loading…</p>
      </div>
    {:else if errorMsg}
      <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
    {:else if booksState.all.length === 0}
      <div
        class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
      >
        <p class="text-muted-foreground">No books yet.</p>
      </div>
    {:else}
      <BookViewControls />
      {#if sortedBooks.length === 0}
        <div
          class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
        >
          <p class="text-muted-foreground">No books match the current filters.</p>
        </div>
      {:else if viewSettings.mode === 'grid'}
        <div
          class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8"
        >
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
        />
      {/if}
    {/if}
  </div>

  {#if filterState.open}
    <BookFilterSidebar books={booksState.all} />
  {/if}
</div>

<SelectionToolbar
  {selectedIds}
  books={booksState.all}
  ondeselect={() => (selectedIds = new Set())}
  onselectall={() => (selectedIds = new Set(sortedBooks.map((b) => b.id)))}
/>
