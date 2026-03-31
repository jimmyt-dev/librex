<script lang="ts">
  import { type Book } from '$lib/api/books.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import BookIcon from '@lucide/svelte/icons/book';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';

  let {
    seriesName,
    books,
    gridClasses,
    selectedIds,
    selectMode,
    onselect
  }: {
    seriesName: string;
    books: Book[];
    gridClasses: string;
    selectedIds: Set<string>;
    selectMode: boolean;
    onselect: (id: string, selected: boolean, shiftKey: boolean) => void;
  } = $props();

  let expanded = $state(true);

  let coverBook = $derived.by(() => {
    const withNumber = books.filter((b) => b.metadata.seriesNumber !== null);
    if (withNumber.length > 0) {
      return withNumber.reduce((min, b) =>
        (b.metadata.seriesNumber ?? Infinity) < (min.metadata.seriesNumber ?? Infinity) ? b : min
      );
    }
    return books[0];
  });
</script>

<div class="flex flex-col gap-2">
  <!-- Header -->
  <button
    type="button"
    class="flex w-full items-center gap-3 rounded-lg border bg-muted/20 px-3 py-2.5 text-left transition-colors hover:bg-muted/40"
    onclick={() => (expanded = !expanded)}
  >
    <!-- Cover thumbnail -->
    <div
      class="flex aspect-[2/3] h-12 shrink-0 items-center justify-center overflow-hidden rounded-md bg-muted"
    >
      {#if coverBook?.metadata.coverPath}
        <img
          src="/api/books/{coverBook.id}/cover"
          alt={seriesName}
          class="h-full w-full object-cover"
          loading="lazy"
        />
      {:else}
        <BookIcon class="size-4 text-muted-foreground" />
      {/if}
    </div>

    <!-- Info -->
    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-semibold">{seriesName}</p>
      <p class="text-xs text-muted-foreground">{books.length} volumes</p>
    </div>

    <!-- Chevron -->
    <ChevronDownIcon
      class="size-4 shrink-0 text-muted-foreground transition-transform {expanded
        ? ''
        : '-rotate-90'}"
    />
  </button>

  <!-- Books grid -->
  {#if expanded}
    <div class={gridClasses}>
      {#each books as book (book.id)}
        <BookCard {book} selected={selectedIds.has(book.id)} {selectMode} {onselect} />
      {/each}
    </div>
  {/if}
</div>
