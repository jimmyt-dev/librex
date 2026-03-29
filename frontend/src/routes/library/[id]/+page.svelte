<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { page } from '$app/state';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import BookView from '$lib/components/book-view.svelte';

  let libraryId = $derived(page.params.id || '');
  let library = $derived(librariesState.items.find((l) => l.id === libraryId));
  let books = $derived(booksState.get(libraryId));

  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

  $effect(() => {
    headerState.title = library?.title ?? 'Library';
    headerState.subtitle = null;
  });

  $effect(() => {
    const id = libraryId;
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
        if (elapsed < 300) await new Promise((r) => setTimeout(r, 300 - elapsed));
        isLoading = false;
      });
  });
</script>

<BookView {books} {isLoading} {errorMsg} emptyMessage="No books yet. Import some from Bookdrop." />
