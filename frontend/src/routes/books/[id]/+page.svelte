<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { apiFetch } from '$lib/api/client';
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { filterState, type ItemState } from '$lib/state/filter.svelte';
  import { SvelteMap } from 'svelte/reactivity';
  import { shelvesState, type Shelf } from '$lib/api/shelves.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import ReadingProgressControls from '$lib/components/reading-progress-controls.svelte';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { Label } from '$lib/components/ui/label';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import { Button } from '$lib/components/ui/button';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { toast } from 'svelte-sonner';
  import BookCard from '$lib/components/book-card.svelte';
  import BookIcon from '@lucide/svelte/icons/book';
  import BookCopyIcon from '@lucide/svelte/icons/book-copy';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';

  let bookId = $derived(page.params.id);

  let book = $state<Book | null>(null);
  let shelves = $state<Shelf[]>([]);
  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);
  let deleteOpen = $state(false);
  let deleteFile = $state(false);
  let deleting = $state(false);
  let descExpanded = $state(false);

  $effect(() => {
    headerState.title = book?.metadata.title ?? 'Book';
    headerState.subtitle = null;
    headerState.counts = [];
  });

  $effect(() => {
    const id = bookId;
    isLoading = true;
    errorMsg = null;
    booksState.fetchAll().catch(() => {});
    Promise.all([
      apiFetch(`/api/books/${id}`),
      apiFetch(`/api/books/${id}/shelves`).catch(() => [])
    ])
      .then(([b, s]) => {
        book = b;
        shelves = s;
        booksState.upsert(b);
      })
      .catch((e: unknown) => {
        errorMsg = e instanceof Error ? e.message : 'Failed to load book.';
      })
      .finally(() => {
        isLoading = false;
      });
  });

  // Keep book in sync with global state (e.g. after edit)
  let stateBook = $derived(book ? booksState.find(book.id) : null);
  $effect(() => {
    if (stateBook) book = stateBook;
  });

  async function download() {
    try {
      const token = localStorage.getItem('bearer_token') || '';
      const res = await fetch(`/api/books/${bookId}/download`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Download failed');
      const blob = await res.blob();
      const match = (res.headers.get('Content-Disposition') || '').match(/filename="(.+?)"/);
      const filename = match ? match[1] : (book?.metadata.title ?? 'book');
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = filename;
      a.click();
      URL.revokeObjectURL(a.href);
    } catch {
      toast.error('Failed to download book.');
    }
  }

  async function confirmDelete() {
    if (!book) return;
    deleting = true;
    try {
      await booksState.delete(book.id, deleteFile);
      toast.success(`"${book.metadata.title}" deleted.`);
      librariesState.fetchAll();
      shelvesState.fetchAll();
      goto('/all-books');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete book.');
    } finally {
      deleting = false;
    }
  }

  function formatDate(d: string | null | undefined): string {
    if (!d) return '—';
    return new Date(d + 'T00:00:00').toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }

  function formatSeries(book: Book): string {
    const { seriesName, seriesNumber } = book.metadata;
    if (!seriesName) return '';
    if (seriesNumber == null) return seriesName;
    const num = Number.isInteger(seriesNumber)
      ? seriesNumber
      : Number(seriesNumber)
          .toFixed(2)
          .replace(/\.?0+$/, '');
    return `${seriesName} #${num}`;
  }

  function formatExt(filePath: string): string {
    return filePath.split('.').pop()?.toUpperCase() ?? '';
  }

  let seriesBooks = $derived.by(() => {
    const name = book?.metadata.seriesName;
    if (!name) return [];
    return booksState.all
      .filter((b) => b.metadata.seriesName === name && b.id !== book?.id)
      .sort(
        (a, b) => (a.metadata.seriesNumber ?? Infinity) - (b.metadata.seriesNumber ?? Infinity)
      );
  });

  function filterByGenre(name: string) {
    filterState.genreSelections = new SvelteMap<string, ItemState>([[name, 'include']]);
    filterState.open = true;
    goto('/all-books');
  }

  function filterByTag(name: string) {
    filterState.tagSelections = new SvelteMap<string, ItemState>([[name, 'include']]);
    filterState.open = true;
    goto('/all-books');
  }
</script>

<div class="page-content gap-6">
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
  {/if}

  {#if isLoading}
    <div class="flex gap-6">
      <Skeleton class="h-64 w-44 shrink-0 rounded-lg" />
      <div class="flex flex-1 flex-col gap-3 pt-1">
        <Skeleton class="h-7 w-3/4" />
        <Skeleton class="h-4 w-1/2" />
        <Skeleton class="h-4 w-1/3" />
        <Skeleton class="h-4 w-1/4" />
      </div>
    </div>
  {:else if book}
    <!-- Top section: cover + metadata -->
    <div class="flex flex-col gap-6 sm:flex-row">
      <!-- Cover -->
      <div class="shrink-0">
        {#if book.metadata.coverPath}
          <img
            src="/api/books/{book.id}/cover"
            alt={book.metadata.title}
            class="h-64 w-44 rounded-lg object-cover shadow-md"
          />
        {:else}
          <div
            class="flex h-64 w-44 items-center justify-center rounded-lg border bg-muted text-muted-foreground shadow-md"
          >
            <BookIcon class="size-12" />
          </div>
        {/if}
      </div>

      <!-- Metadata -->
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <div>
          <h1 class="text-2xl leading-tight font-bold">{book.metadata.title}</h1>
          {#if book.metadata.subtitle}
            <p class="mt-0.5 text-base text-muted-foreground">{book.metadata.subtitle}</p>
          {/if}
        </div>

        {#if book.authors.length > 0}
          <p class="text-sm">
            {#each book.authors as author, i (author.id)}
              <a href="/authors/{author.id}" class="font-medium hover:underline">{author.name}</a
              >{#if i < book.authors.length - 1},
              {/if}
            {/each}
          </p>
        {/if}

        {#if formatSeries(book)}
          <p class="text-sm text-muted-foreground">{formatSeries(book)}</p>
        {/if}

        <!-- Genres & Tags -->
        {#if book.genres.length > 0 || book.tags.length > 0}
          <div class="flex flex-wrap gap-1.5">
            {#each book.genres as genre (genre.id)}
              <button
                type="button"
                onclick={() => filterByGenre(genre.name)}
                class="cursor-pointer rounded-full border bg-muted px-2.5 py-0.5 text-xs font-medium hover:bg-muted/80"
              >
                {genre.name}
              </button>
            {/each}
            {#each book.tags as tag (tag.id)}
              <button
                type="button"
                onclick={() => filterByTag(tag.name)}
                class="cursor-pointer rounded-full border px-2.5 py-0.5 text-xs text-muted-foreground hover:bg-muted/80"
              >
                {tag.name}
              </button>
            {/each}
          </div>
        {/if}

        <!-- Metadata grid -->
        <div class="grid grid-cols-2 gap-x-8 gap-y-1 text-sm sm:grid-cols-3">
          {#if book.metadata.publisher}
            <div>
              <span class="text-muted-foreground">Publisher</span>
              <p class="font-medium">{book.metadata.publisher}</p>
            </div>
          {/if}
          {#if book.metadata.publishedDate}
            <div>
              <span class="text-muted-foreground">Published</span>
              <p class="font-medium">{formatDate(book.metadata.publishedDate)}</p>
            </div>
          {/if}
          {#if book.metadata.language}
            <div>
              <span class="text-muted-foreground">Language</span>
              <p class="font-medium">{book.metadata.language}</p>
            </div>
          {/if}
          {#if book.metadata.pageCount}
            <div>
              <span class="text-muted-foreground">Pages</span>
              <p class="font-medium">{book.metadata.pageCount}</p>
            </div>
          {/if}
          {#if book.metadata.isbn13 ?? book.metadata.isbn10}
            <div>
              <span class="text-muted-foreground">ISBN</span>
              <p class="font-mono text-xs font-medium">
                {book.metadata.isbn13 ?? book.metadata.isbn10}
              </p>
            </div>
          {/if}
          {#if book.metadata.rating}
            <div>
              <span class="text-muted-foreground">Rating</span>
              <p class="text-yellow-400">
                {'★'.repeat(book.metadata.rating)}<span class="text-muted-foreground/30"
                  >{'★'.repeat(5 - book.metadata.rating)}</span
                >
              </p>
            </div>
          {/if}
          <div>
            <span class="text-muted-foreground">Format</span>
            <p class="font-medium">{formatExt(book.filePath)}</p>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="mt-1 flex flex-wrap gap-2">
          <Button variant="outline" size="sm" onclick={() => bookEditState.openFor(book!)}>
            <PencilIcon class="size-3.5" /> Edit
          </Button>
          <Button variant="outline" size="sm" onclick={download}>
            <DownloadIcon class="size-3.5" /> Download
          </Button>
          <Button
            variant="outline"
            size="sm"
            class="text-destructive hover:text-destructive"
            onclick={() => (deleteOpen = true)}
          >
            <TrashIcon class="size-3.5" /> Delete
          </Button>
        </div>
      </div>
    </div>

    <!-- Reading Progress -->
    <div class="rounded-lg border p-4">
      <h2 class="mb-4 text-sm font-semibold">Reading Progress</h2>
      <ReadingProgressControls {book} />
    </div>

    <!-- Description -->
    {#if book.metadata.description}
      <div class="rounded-lg border p-4">
        <h2 class="mb-2 text-sm font-semibold">Description</h2>
        <div class="relative">
          <p
            class="text-sm leading-relaxed text-muted-foreground {descExpanded
              ? ''
              : 'line-clamp-4'}"
          >
            {book.metadata.description}
          </p>
          {#if book.metadata.description.length > 300}
            <button
              type="button"
              class="mt-1 text-xs text-primary hover:underline"
              onclick={() => (descExpanded = !descExpanded)}
            >
              {descExpanded ? 'Show less' : 'Show more'}
            </button>
          {/if}
        </div>
      </div>
    {/if}

    <!-- Shelves -->
    {#if shelves.length > 0}
      <div class="rounded-lg border p-4">
        <h2 class="mb-3 flex items-center gap-1.5 text-sm font-semibold">
          <LibraryBigIcon class="size-3.5" /> On Shelves
        </h2>
        <div class="flex flex-wrap gap-2">
          {#each shelves as shelf (shelf.id)}
            <a
              href="/shelf/{shelf.id}"
              class="rounded-full border bg-muted px-3 py-1 text-sm font-medium hover:bg-muted/80"
            >
              {shelf.title}
            </a>
          {/each}
        </div>
      </div>
    {/if}

    <!-- More in series -->
    {#if seriesBooks.length > 0}
      <div class="rounded-lg border p-4">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="flex items-center gap-1.5 text-sm font-semibold">
            <BookCopyIcon class="size-3.5" />
            More in {book.metadata.seriesName}
          </h2>
          <a
            href="/series/{encodeURIComponent(book.metadata.seriesName ?? '')}"
            class="text-xs text-muted-foreground hover:text-foreground hover:underline"
          >
            View all
          </a>
        </div>
        <div class="flex gap-4 overflow-x-auto pb-2">
          {#each seriesBooks as sb (sb.id)}
            <div class="w-28 shrink-0">
              <BookCard book={sb} checkboxes={false} />
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

<AlertDialog.Root
  open={deleteOpen}
  onOpenChange={(o) => {
    if (!o) {
      deleteOpen = false;
      deleteFile = false;
    }
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete "{book?.metadata.title}"?</AlertDialog.Title>
      <AlertDialog.Description>
        This will remove the book from your library. This action cannot be undone.
      </AlertDialog.Description>
    </AlertDialog.Header>
    <Label class="flex cursor-pointer items-center gap-2 text-sm">
      <Checkbox bind:checked={deleteFile} />
      Also delete the file from disk
    </Label>
    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
