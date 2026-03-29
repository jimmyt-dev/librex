<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { authorsState } from '$lib/api/authors.svelte';
  import BookCard from '$lib/components/book-card.svelte';
  import BookIcon from '@lucide/svelte/icons/book';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import UserIcon from '@lucide/svelte/icons/user';
  import BookOpenIcon from '@lucide/svelte/icons/book-open';

  headerState.title = 'Dashboard';
  headerState.subtitle = null;
  headerState.counts = [];

  let isLoading = $state(true);

  $effect(() => {
    Promise.all([booksState.fetchAll(), authorsState.fetchAll()]).finally(() => {
      isLoading = false;
    });
  });

  let totalBooks = $derived(booksState.all.length);
  let totalLibraries = $derived(librariesState.items.length);
  let totalAuthors = $derived(authorsState.items.length);

  let currentlyReading = $derived(
    booksState.all.filter((b) => b.progress?.status === 'reading')
  );

  let recentlyAdded = $derived(
    [...booksState.all].sort((a, b) => new Date(b.addedOn).getTime() - new Date(a.addedOn).getTime()).slice(0, 12)
  );
</script>

<div class="flex flex-1 flex-col gap-6 p-4 pt-0">
  <!-- Stats row -->
  <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
    <div class="flex flex-col gap-1 rounded-lg border bg-card p-4 shadow-sm">
      <div class="flex items-center gap-2 text-muted-foreground">
        <BookIcon class="size-4" />
        <span class="text-xs font-medium uppercase tracking-wide">Books</span>
      </div>
      <p class="text-2xl font-bold">{isLoading ? '—' : totalBooks}</p>
    </div>
    <div class="flex flex-col gap-1 rounded-lg border bg-card p-4 shadow-sm">
      <div class="flex items-center gap-2 text-muted-foreground">
        <LibraryBigIcon class="size-4" />
        <span class="text-xs font-medium uppercase tracking-wide">Libraries</span>
      </div>
      <p class="text-2xl font-bold">{totalLibraries}</p>
    </div>
    <div class="flex flex-col gap-1 rounded-lg border bg-card p-4 shadow-sm">
      <div class="flex items-center gap-2 text-muted-foreground">
        <UserIcon class="size-4" />
        <span class="text-xs font-medium uppercase tracking-wide">Authors</span>
      </div>
      <p class="text-2xl font-bold">{isLoading ? '—' : totalAuthors}</p>
    </div>
    <div class="flex flex-col gap-1 rounded-lg border bg-card p-4 shadow-sm">
      <div class="flex items-center gap-2 text-muted-foreground">
        <BookOpenIcon class="size-4" />
        <span class="text-xs font-medium uppercase tracking-wide">Reading</span>
      </div>
      <p class="text-2xl font-bold">{isLoading ? '—' : currentlyReading.length}</p>
    </div>
  </div>

  {#if !isLoading && totalBooks === 0}
    <div
      class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <div class="text-center">
        <BookIcon class="mx-auto mb-2 size-10 text-muted-foreground/40" />
        <p class="font-medium text-muted-foreground">Your library is empty</p>
        <p class="mt-1 text-sm text-muted-foreground/70">
          <a href="/settings" class="text-primary hover:underline">Add a library</a> or drop books in
          <a href="/bookdrop" class="text-primary hover:underline">Bookdrop</a>
        </p>
      </div>
    </div>
  {:else}
    <!-- Currently Reading -->
    {#if !isLoading && currentlyReading.length > 0}
      <div class="flex flex-col gap-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold">Currently Reading</h2>
        </div>
        <div class="flex gap-4 overflow-x-auto pb-2">
          {#each currentlyReading as book (book.id)}
            <div class="w-32 shrink-0">
              <BookCard {book} checkboxes={false} />
              {#if book.progress?.progress}
                <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-muted">
                  <div
                    class="h-full rounded-full bg-blue-500"
                    style="width: {book.progress.progress}%"
                  ></div>
                </div>
                <p class="mt-0.5 text-center text-[10px] text-muted-foreground">
                  {Math.round(book.progress.progress)}%
                </p>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Recently Added -->
    {#if recentlyAdded.length > 0}
      <div class="flex flex-col gap-3">
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold">Recently Added</h2>
          <a href="/all-books" class="text-xs text-muted-foreground hover:text-foreground hover:underline">
            View all
          </a>
        </div>
        {#if isLoading}
          <div class="flex gap-4 overflow-x-auto pb-2">
            {#each Array(8) as _, i (i)}
              <div class="w-32 shrink-0 rounded-md bg-muted/50 aspect-2/3"></div>
            {/each}
          </div>
        {:else}
          <div class="flex gap-4 overflow-x-auto pb-2">
            {#each recentlyAdded as book (book.id)}
              <div class="w-32 shrink-0">
                <BookCard {book} checkboxes={false} />
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>
