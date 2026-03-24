<script lang="ts">
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Separator } from '$lib/components/ui/separator';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { page } from '$app/state';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState, type Book } from '$lib/api/books.svelte';

  let libraryId = $derived(page.params.id);
  let library = $derived(librariesState.items.find((l) => l.id === libraryId));
  let books = $derived(booksState.get(libraryId));

  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);

  // Edit sheet
  let sheetOpen = $state(false);
  let editingBook = $state<Book | null>(null);
  let editTitle = $state('');
  let editAuthor = $state('');
  let editSubject = $state('');
  let editDescription = $state('');
  let editPublisher = $state('');
  let editContributor = $state('');
  let editDate = $state('');
  let editIdentifier = $state('');
  let editLanguage = $state('');
  let isSaving = $state(false);

  $effect(() => {
    const id = libraryId;
    // Re-runs when libraryId changes OR when cache is invalidated (has() becomes false)
    if (booksState.has(id)) return;
    isLoading = true;
    errorMsg = null;
    booksState.fetchForLibrary(id).catch((e: unknown) => {
      errorMsg = e instanceof Error ? e.message : 'Failed to load books.';
    }).finally(() => {
      isLoading = false;
    });
  });

  function getExt(filePath: string) {
    const dot = filePath.lastIndexOf('.');
    return dot !== -1 ? filePath.slice(dot + 1).toUpperCase() : '?';
  }

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

  function openEdit(book: Book) {
    editingBook = book;
    editTitle = book.title;
    editAuthor = book.author ?? '';
    editSubject = book.subject ?? '';
    editDescription = book.description ?? '';
    editPublisher = book.publisher ?? '';
    editContributor = book.contributor ?? '';
    editDate = book.date ?? '';
    editIdentifier = book.identifier ?? '';
    editLanguage = book.language ?? '';
    sheetOpen = true;
  }

  async function saveEdit() {
    if (!editingBook) return;
    isSaving = true;
    try {
      const res = await fetch(`/api/books/${editingBook.id}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: editTitle,
          author: editAuthor || null,
          subject: editSubject || null,
          description: editDescription || null,
          publisher: editPublisher || null,
          contributor: editContributor || null,
          date: editDate || null,
          identifier: editIdentifier || null,
          language: editLanguage || null
        })
      });
      if (!res.ok) throw new Error('Failed to save');
      const updated: Book = await res.json();
      booksState.upsert(updated);
      sheetOpen = false;
    } catch {
      errorMsg = 'Failed to save changes.';
    } finally {
      isSaving = false;
    }
  }
</script>

<header
  class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
>
  <div class="flex w-full items-center gap-2 px-4">
    <Sidebar.Trigger class="-ms-1" />
    <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
    <Breadcrumb.Root>
      <Breadcrumb.List>
        <Breadcrumb.Item>
          <Breadcrumb.Page>{library?.title ?? 'Library'}</Breadcrumb.Page>
        </Breadcrumb.Item>
      </Breadcrumb.List>
    </Breadcrumb.Root>
    {#if !isLoading}
      <span class="text-sm text-muted-foreground">
        {books.length} book{books.length === 1 ? '' : 's'}
      </span>
    {/if}
  </div>
</header>

<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
  {/if}

  {#if isLoading}
    <div class="flex min-h-64 items-center justify-center">
      <p class="text-muted-foreground">Loading…</p>
    </div>
  {:else if books.length === 0}
    <div
      class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <p class="text-muted-foreground">No books yet. Import some from Bookdrop.</p>
    </div>
  {:else}
    <div
      class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6"
    >
      {#each books as book (book.id)}
        <button
          class="group flex flex-col gap-2 rounded-lg border bg-card p-3 text-left transition-colors hover:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onclick={() => openEdit(book)}
        >
          <div class="flex aspect-2/3 w-full items-center justify-center overflow-hidden rounded bg-muted/50">
            {#if book.cover}
              <img src={book.cover} alt={book.title} class="h-full w-full object-cover" />
            {:else}
              <span class="text-2xl font-bold text-muted-foreground/40">{getExt(book.filePath)}</span>
            {/if}
          </div>
          <div class="flex flex-col gap-0.5">
            <p class="line-clamp-2 text-sm font-medium leading-tight" title={book.title}>
              {book.title}
            </p>
            <p class="truncate text-xs text-muted-foreground" title={book.author ?? undefined}>
              {book.author ?? '—'}
            </p>
          </div>
        </button>
      {/each}
    </div>
  {/if}
</div>

<Sheet.Root bind:open={sheetOpen}>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      {#if editingBook}
        <Sheet.Header>
          <Sheet.Title>Edit Metadata</Sheet.Title>
          <Sheet.Description class="truncate text-xs text-muted-foreground">
            {editingBook.filePath.split('/').pop()}
          </Sheet.Description>
        </Sheet.Header>
        <div class="flex flex-col gap-4 overflow-y-auto px-4 py-6">
          {#if editingBook.cover}
            <img
              src={editingBook.cover}
              alt="Cover"
              class="mx-auto h-48 w-auto rounded object-contain shadow"
            />
          {/if}
          <div class="flex flex-col gap-1.5">
            <label for="edit-title" class="text-sm font-medium">Title</label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-author" class="text-sm font-medium">Author</label>
            <Input id="edit-author" bind:value={editAuthor} placeholder="Unknown" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-subject" class="text-sm font-medium">Subject</label>
            <Input id="edit-subject" bind:value={editSubject} placeholder="Genre / topics" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-description" class="text-sm font-medium">Description</label>
            <Input id="edit-description" bind:value={editDescription} placeholder="Synopsis" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-publisher" class="text-sm font-medium">Publisher</label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-contributor" class="text-sm font-medium">Contributor</label>
            <Input id="edit-contributor" bind:value={editContributor} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <label for="edit-date" class="text-sm font-medium">Date</label>
              <Input id="edit-date" bind:value={editDate} placeholder="YYYY" />
            </div>
            <div class="flex flex-col gap-1.5">
              <label for="edit-language" class="text-sm font-medium">Language</label>
              <Input id="edit-language" bind:value={editLanguage} placeholder="en" />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-identifier" class="text-sm font-medium">Identifier (ISBN)</label>
            <Input id="edit-identifier" bind:value={editIdentifier} />
          </div>
        </div>
        <Sheet.Footer>
          <Sheet.Close>
            {#snippet child({ props })}
              <Button variant="outline" {...props}>Cancel</Button>
            {/snippet}
          </Sheet.Close>
          <Button onclick={saveEdit} disabled={isSaving || !editTitle.trim()}>
            {isSaving ? 'Saving…' : 'Save'}
          </Button>
        </Sheet.Footer>
      {/if}
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
