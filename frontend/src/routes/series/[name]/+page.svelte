<script lang="ts">
  import { page } from '$app/state';
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import BookView from '$lib/components/book-view.svelte';

  let seriesName = $derived(decodeURIComponent(page.params.name ?? ''));

  $effect(() => {
    headerState.title = seriesName;
    headerState.subtitle = null;
  });

  let isLoading = $state(true);

  $effect(() => {
    isLoading = true;
    booksState.fetchAll().finally(() => {
      isLoading = false;
    });
  });

  let books = $derived(booksState.all.filter((b) => b.metadata.seriesName === seriesName));
</script>

<BookView {books} {isLoading} emptyMessage="No books found in this series." />
