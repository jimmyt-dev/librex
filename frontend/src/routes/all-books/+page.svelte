<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import BookView from '$lib/components/book-view.svelte';

  headerState.title = 'All Books';
  headerState.subtitle = null;

  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

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
</script>

<BookView books={booksState.all} {isLoading} {errorMsg} emptyMessage="No books yet." />
