<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import SearchIcon from '@lucide/svelte/icons/search';
  import BookIcon from '@lucide/svelte/icons/book';
  import * as InputGroup from '$lib/components/ui/input-group';
  import * as Kbd from '$lib/components/ui/kbd';

  let { autofocus = false }: { autofocus?: boolean } = $props();

  let query = $state('');
  let focused = $state(false);
  let mousedownOnResult = $state(false);
  let inputEl = $state<HTMLInputElement | null>(null);
  let highlightedIndex = $state(-1);

  let isMac = $derived(typeof navigator !== 'undefined' && navigator.platform.startsWith('Mac'));

  $effect(() => {
    if (autofocus) {
      setTimeout(() => inputEl?.focus(), 0);
    }
  });

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

  $effect(() => {
    void results;
    highlightedIndex = -1;
  });

  $effect(() => {
    void page.url.pathname;
    query = '';
    focused = false;
    highlightedIndex = -1;
  });

  function selectBook(book: Book) {
    query = '';
    focused = false;
    highlightedIndex = -1;
    goto(`/books/${book.id}`);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      query = '';
      inputEl?.blur();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (results.length > 0) highlightedIndex = (highlightedIndex + 1) % results.length;
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (results.length > 0)
        highlightedIndex = (highlightedIndex - 1 + results.length) % results.length;
    } else if (e.key === 'Enter' && highlightedIndex >= 0 && results[highlightedIndex]) {
      e.preventDefault();
      selectBook(results[highlightedIndex]);
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
  class="relative transition-[width] duration-200 ease-in-out {autofocus ? 'w-full' : ''}"
  class:w-60={!autofocus && !focused}
  class:w-80={!autofocus && focused}
>
  <InputGroup.Root class="h-9">
    <InputGroup.Addon align="inline-start">
      <SearchIcon class="size-3.5" />
    </InputGroup.Addon>

    <InputGroup.Input
      bind:ref={inputEl}
      bind:value={query}
      placeholder="Search books..."
      onfocus={() => (focused = true)}
      onblur={() => {
        if (!mousedownOnResult) focused = false;
      }}
      onkeydown={handleKeydown}
    />

    {#if !focused}
      <InputGroup.Addon align="inline-end" class="hidden sm:block">
        <Kbd.Group>
          <Kbd.Root>{isMac ? '⌘' : 'Ctrl'}</Kbd.Root>
          <Kbd.Root>K</Kbd.Root>
        </Kbd.Group>
      </InputGroup.Addon>
    {/if}
  </InputGroup.Root>

  {#if showDropdown}
    <div
      class="absolute top-full left-0 z-50 mt-1.5 w-full min-w-80 overflow-hidden rounded-lg border bg-popover shadow-lg"
    >
      {#if results.length === 0}
        <p class="px-3 py-2.5 text-sm text-muted-foreground">No results found.</p>
      {:else}
        {#each results as book, i (book.id)}
          <div
            class="flex cursor-pointer items-center gap-2.5 px-3 py-2 hover:bg-accent {i ===
            highlightedIndex
              ? 'bg-accent'
              : ''}"
            role="button"
            tabindex="0"
            onmousedown={() => (mousedownOnResult = true)}
            onmouseup={() => (mousedownOnResult = false)}
            onmouseenter={() => (highlightedIndex = i)}
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
