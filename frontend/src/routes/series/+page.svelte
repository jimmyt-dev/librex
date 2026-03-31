<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import BookIcon from '@lucide/svelte/icons/book';
  import { SvelteMap } from 'svelte/reactivity';

  headerState.title = 'Series';
  headerState.subtitle = null;

  let isLoading = $state(true);

  $effect(() => {
    isLoading = true;
    booksState.fetchAll().finally(() => {
      isLoading = false;
    });
  });

  type SeriesEntry = {
    name: string;
    coverBook: Book;
    count: number;
  };

  let seriesList = $derived.by((): SeriesEntry[] => {
    const map = new SvelteMap<string, Book[]>();
    for (const book of booksState.all) {
      const name = book.metadata.seriesName;
      if (!name) continue;
      const existing = map.get(name);
      if (existing) {
        existing.push(book);
      } else {
        map.set(name, [book]);
      }
    }
    const entries: SeriesEntry[] = [];
    for (const [name, books] of map) {
      const sorted = [...books].sort((a, b) => {
        const an = a.metadata.seriesNumber ?? Infinity;
        const bn = b.metadata.seriesNumber ?? Infinity;
        return an - bn;
      });
      entries.push({ name, coverBook: sorted[0], count: books.length });
    }
    entries.sort((a, b) => a.name.localeCompare(b.name));
    return entries;
  });
</script>

<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
  {#if isLoading}
    <div
      class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8"
    >
      <!-- eslint-disable-next-line @typescript-eslint/no-unused-vars -->
      {#each Array(12) as _, i (i)}
        <div class="flex flex-col gap-2">
          <div class="aspect-2/3 animate-pulse rounded-lg bg-muted"></div>
          <div class="h-4 w-3/4 animate-pulse rounded bg-muted"></div>
          <div class="h-3 w-1/2 animate-pulse rounded bg-muted"></div>
        </div>
      {/each}
    </div>
  {:else if seriesList.length === 0}
    <div
      class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <p class="text-muted-foreground">No series found.</p>
    </div>
  {:else}
    <div
      class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8"
    >
      {#each seriesList as series (series.name)}
        <a href="/series/{encodeURIComponent(series.name)}" class="flex flex-col gap-2">
          <div class="aspect-2/3 overflow-hidden rounded-lg">
            {#if series.coverBook.metadata.coverPath}
              <img
                src="/api/books/{series.coverBook.id}/cover"
                alt={series.name}
                class="h-full w-full object-cover"
              />
            {:else}
              <div
                class="flex h-full w-full items-center justify-center bg-muted text-muted-foreground"
              >
                <BookIcon class="size-10" />
              </div>
            {/if}
          </div>
          <div>
            <p class="truncate leading-tight font-medium">{series.name}</p>
            <p class="text-xs text-muted-foreground">
              {series.count}
              {series.count === 1 ? 'book' : 'books'}
            </p>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>
