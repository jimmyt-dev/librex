<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { page } from '$app/state';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import SelectionToolbar from '$lib/components/selection-toolbar.svelte';
  import { SvelteSet } from 'svelte/reactivity';

  let shelfId = $derived(page.params.id || '');
  let shelf = $derived(
    shelfId === 'unshelved'
      ? { id: 'unshelved', title: 'Unshelved', books: shelvesState.unshelvedCount }
      : shelvesState.items.find((s) => s.id === shelfId)
  );
  let books = $derived(shelvesState.get(shelfId));

  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);
  let selectedIds = $state<Set<string>>(new Set());
  let lastSelectedId = $state<string | null>(null);

  $effect(() => {
    const bookIds = new Set(books.map((b) => b.id));
    const pruned = new SvelteSet([...selectedIds].filter((id) => bookIds.has(id)));
    if (pruned.size !== selectedIds.size) selectedIds = pruned;
  });

  $effect(() => {
    headerState.title = shelf?.title ?? 'Shelf';
    headerState.subtitle = null;
    headerState.counts = isLoading ? [] : [{ label: 'books', value: books.length }];
  });

  $effect(() => {
    const id = shelfId;
    if (shelvesState.has(id)) return;
    isLoading = true;
    errorMsg = null;
    shelvesState
      .fetchBooksForShelf(id)
      .catch((e: unknown) => {
        errorMsg = e instanceof Error ? e.message : 'Failed to load books.';
      })
      .finally(() => {
        isLoading = false;
      });
  });

  function toggleSelect(id: string, sel: boolean, shiftKey: boolean) {
    if (shiftKey && sel && lastSelectedId) {
      const ids = books.map((b) => b.id);
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

<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
  {/if}

  {#if isLoading}
    <div class="flex min-h-64 items-center justify-center">
      <p class="text-muted-foreground">Loading…</p>
    </div>
  {:else if books.length === 0}
    <div
      class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <p class="text-muted-foreground">
        {shelfId === 'unshelved' ? 'All books are shelved.' : 'No books on this shelf yet.'}
      </p>
    </div>
  {:else}
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
      {#each books as book (book.id)}
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
  {books}
  ondeselect={() => (selectedIds = new Set())}
  onselectall={() => (selectedIds = new Set(books.map((b) => b.id)))}
/>
