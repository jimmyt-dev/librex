<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { page } from '$app/state';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { viewSettings } from '$lib/state/view-settings.svelte';
  import { filterState } from '$lib/state/filter.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import BookCardSkeleton from '$lib/components/book-card-skeleton.svelte';
  import BookTable from '$lib/components/book-table.svelte';
  import BookViewControls from '$lib/components/book-view-controls.svelte';
  import BookFilterSidebar from '$lib/components/book-filter-sidebar.svelte';
  import SelectionToolbar from '$lib/components/selection-toolbar.svelte';
  import { SvelteSet } from 'svelte/reactivity';

  let libraryId = $derived(page.params.id || '');
  let library = $derived(librariesState.items.find((l) => l.id === libraryId));
  let books = $derived(booksState.get(libraryId));
  let sortedBooks = $derived(viewSettings.sort(filterState.apply(books)));

  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);
  let selectedIds = $state<Set<string>>(new Set());
  let lastSelectedId = $state<string | null>(null);

  $effect(() => {
    const bookIds = new Set(books.map((b) => b.id));
    const pruned = new SvelteSet([...selectedIds].filter((id) => bookIds.has(id)));
    if (pruned.size !== selectedIds.size) selectedIds = pruned;
  });

  $effect(() => {
    headerState.title = library?.title ?? 'Library';
    headerState.subtitle = null;
    headerState.counts = isLoading ? [] : [{ label: 'books', value: sortedBooks.length }];
  });
$effect(() => {
  const id = libraryId;
  if (books.length > 0) {
    isLoading = false;
    return;
  }
  const start = Date.now();
  isLoading = true;
  errorMsg = null;
  booksState
    .fetchForLibrary(id)
    .catch((e: unknown) => {
      errorMsg = e instanceof Error ? e.message : 'Failed to load books.';
    })
    .finally(async () => {
      const elapsed = Date.now() - start;
      if (elapsed < 300) {
        await new Promise((r) => setTimeout(r, 300 - elapsed));
      }
      isLoading = false;
    });
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
      <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
    {/if}

    {#if isLoading}
      <div
        class="flex h-12.5 w-full items-center justify-between gap-2 rounded-md border bg-muted/20 p-2"
      ></div>
      <div
        class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8"
      >
        <!-- eslint-disable-next-line @typescript-eslint/no-unused-vars -->
        {#each Array(12) as _, i (i)}
          <BookCardSkeleton />
        {/each}
      </div>
    {:else if library?.books === 0 && !isLoading}
      <div
        class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
      >
        <p class="text-muted-foreground">No books yet. Import some from Bookdrop.</p>
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
          loading={isLoading}
        />
      {/if}
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
