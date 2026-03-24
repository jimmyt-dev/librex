<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import SelectionToolbar from '$lib/components/selection-toolbar.svelte';
  import { SvelteSet } from 'svelte/reactivity';

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
    headerState.counts = isLoading
      ? []
      : [{ label: 'books', value: booksState.all.length }];
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

<div class="px-4">
  {#if isLoading}
    <p>Loading...</p>
  {:else if errorMsg}
    <p>{errorMsg}</p>
  {:else}
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
      {#each booksState.all as book (book.id)}
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

<SelectionToolbar
  {selectedIds}
  books={booksState.all}
  ondeselect={() => (selectedIds = new Set())}
  onselectall={() => (selectedIds = new Set(booksState.all.map((b) => b.id)))}
/>
