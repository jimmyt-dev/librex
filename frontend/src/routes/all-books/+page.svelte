<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';

  let isLoading = $state(false);
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

<div class="p-4">
  {#if isLoading}
    <p>Loading...</p>
  {:else if errorMsg}
    <p>{errorMsg}</p>
  {:else}
    <ul>
      {#each booksState.all as book (book.id)}
        <li>{book.title}</li>
      {/each}
    </ul>
  {/if}
</div>
