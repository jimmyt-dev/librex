<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { goto } from '$app/navigation';
  import SearchIcon from '@lucide/svelte/icons/search';
  import BookIcon from '@lucide/svelte/icons/book';

  let query = $state('');
  let focused = $state(false);
  let mousedownOnResult = $state(false);
  let inputEl = $state<HTMLInputElement | null>(null);

  let isMac = $derived(typeof navigator !== 'undefined' && navigator.platform.startsWith('Mac'));
  let kbdHint = $derived(isMac ? '⌘K' : 'Ctrl K');

  let results = $derived.by((): Book[] => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    const matches: Book[] = [];
    for (const book of booksState.all) {
      if (matches.length >= 8) break;
      const title = (book.metadata.title ?? '').toLowerCase();
      const series = (book.metadata.seriesName ?? '').toLowerCase();
      const authors = book.authors.map((a: { name: string }) => a.name.toLowerCase()).join(' ');
      const genres = book.genres.map((g: { name: string }) => g.name.toLowerCase()).join(' ');
      const isbn = (
        (book.metadata.isbn13 ?? '') +
        ' ' +
        (book.metadata.isbn10 ?? '')
      ).toLowerCase();
      if (
        title.includes(q) ||
        series.includes(q) ||
        authors.includes(q) ||
        genres.includes(q) ||
        isbn.includes(q)
      ) {
        matches.push(book);
      }
    }
    return matches;
  });

  let showDropdown = $derived(focused && query.trim().length > 0);

  function selectBook(book: Book) {
    query = '';
    focused = false;
    goto(`/books/${book.id}`);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      query = '';
      inputEl?.blur();
    }
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      inputEl?.focus();
      inputEl?.select();
    }
  }
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<div
  class="relative transition-[width] duration-200 ease-in-out"
  class:w-80={!focused}
  class:w-100={focused}
>
  <SearchIcon
    class="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground"
  />
  <input
    bind:this={inputEl}
    bind:value={query}
    onfocus={() => (focused = true)}
    onblur={() => {
      if (!mousedownOnResult) focused = false;
    }}
    onkeydown={handleKeydown}
    type="search"
    placeholder="Search by title, author, genre, ISBN..."
    class="h-9 w-full rounded-md border bg-muted/40 pr-14 pl-8 text-sm placeholder:text-muted-foreground/60 focus:bg-background focus:ring-1 focus:ring-ring focus:outline-none"
  />
  {#if !focused}
    <span
      class="pointer-events-none absolute top-1/2 right-2.5 -translate-y-1/2 rounded border bg-muted px-1 py-0.5 text-[10px] font-medium text-muted-foreground"
    >
      {kbdHint}
    </span>
  {/if}

  {#if showDropdown}
    <div
      class="absolute top-full left-0 z-50 mt-1.5 w-full min-w-80 overflow-hidden rounded-lg border bg-popover shadow-lg"
    >
      {#if results.length === 0}
        <p class="px-3 py-2.5 text-sm text-muted-foreground">No results found.</p>
      {:else}
        {#each results as book (book.id)}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="flex cursor-pointer items-center gap-2.5 px-3 py-2 hover:bg-accent"
            role="button"
            tabindex="0"
            onmousedown={() => (mousedownOnResult = true)}
            onmouseup={() => (mousedownOnResult = false)}
            onclick={() => selectBook(book)}
            onkeydown={(e) => e.key === 'Enter' && selectBook(book)}
          >
            <div
              class="flex size-8 shrink-0 items-center justify-center overflow-hidden rounded bg-muted"
            >
              {#if book.metadata.coverPath}
                <img
                  src="/api/books/{book.id}/cover"
                  alt={book.metadata.title}
                  class="h-full w-full object-cover"
                  loading="lazy"
                />
              {:else}
                <BookIcon class="size-3.5 text-muted-foreground" />
              {/if}
            </div>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm leading-tight font-medium">{book.metadata.title}</p>
              {#if book.authors.length > 0}
                <p class="truncate text-xs text-muted-foreground">
                  {book.authors.map((a) => a.name).join(', ')}
                </p>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>
