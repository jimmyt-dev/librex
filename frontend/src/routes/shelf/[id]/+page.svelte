<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { page } from '$app/state';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import BookView from '$lib/components/book-view.svelte';

  let shelfId = $derived(page.params.id || '');
  let shelf = $derived(shelvesState.items.find((s) => s.id === shelfId));
  let books = $derived(shelvesState.get(shelfId));

  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

  $effect(() => {
    headerState.title = shelf?.title ?? 'Shelf';
    headerState.subtitle = null;
  });

  $effect(() => {
    const id = shelfId;
    if (shelvesState.has(id)) {
      isLoading = false;
      return;
    }
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
</script>

<BookView
  {books}
  {isLoading}
  {errorMsg}
  emptyMessage={shelfId === 'unshelved' ? 'All books are shelved.' : 'No books on this shelf yet.'}
/>
