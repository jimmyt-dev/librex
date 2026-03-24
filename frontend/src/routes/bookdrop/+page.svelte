<script lang="ts">
  import * as Sheet from '$lib/components/ui/sheet';
  import { headerState } from '$lib/state/header.svelte';
  headerState.title = 'Bookdrop';
  headerState.subtitle = null;
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { SvelteMap, SvelteSet } from 'svelte/reactivity';
  import * as Select from '$lib/components/ui/select';
  import RotateCW from '@lucide/svelte/icons/rotate-cw';
  import { Spinner } from '$lib/components/ui/spinner';
  import { Checkbox } from '$lib/components/ui/checkbox';

  interface StagedBook {
    id: string;
    title: string;
    author: string | null;
    subject: string | null;
    description: string | null;
    publisher: string | null;
    contributor: string | null;
    date: string | null;
    type: string | null;
    format: string | null;
    identifier: string | null;
    source: string | null;
    language: string | null;
    relation: string | null;
    coverage: string | null;
    hasCover: boolean;
    fileName: string;
    ext: string;
    originalPath: string;
    userId: string;
  }

  function coverUrl(id: string) {
    return `/api/bookdrop/staged/${id}/cover`;
  }

  let stagedBooks = $state<StagedBook[]>([]);
  let isScanning = $state(false);
  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

  // Selection
  let selectedIds = $state<Set<string>>(new Set());
  let allSelected = $derived(stagedBooks.length > 0 && selectedIds.size === stagedBooks.length);

  // Single-book sheet edit
  let sheetOpen = $state(false);
  let editingBook = $state<StagedBook | null>(null);
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

  // Per-book library selection
  let bookLibraryMap = $state<Map<string, string>>(new Map());

  function getBookLibraryTitle(bookId: string) {
    const libId = bookLibraryMap.get(bookId);
    return libId
      ? (librariesState.items.find((l) => l.id === libId)?.title ?? 'Library…')
      : 'Library…';
  }

  // How many selected books have a library assigned
  let readyToImportCount = $derived([...selectedIds].filter((id) => bookLibraryMap.has(id)).length);

  let isImporting = $state(false);

  async function handleAddAllToLibraries() {
    const items = [...selectedIds]
      .filter((id) => bookLibraryMap.has(id))
      .map((id) => ({ stagedBookId: id, libraryId: bookLibraryMap.get(id)! }));

    if (items.length === 0) return;

    isImporting = true;
    errorMsg = null;
    try {
      const res = await fetch('/api/bookdrop/import', {
        method: 'POST',
        headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'application/json' },
        body: JSON.stringify(items)
      });
      if (!res.ok) throw new Error('Import request failed');

      const results: { stagedBookId: string; error?: string }[] = await res.json();

      const succeeded = new Set(results.filter((r) => !r.error).map((r) => r.stagedBookId));
      const failed = results.filter((r) => r.error);

      // Remove successfully imported books from local state
      stagedBooks = stagedBooks.filter((b) => !succeeded.has(b.id));
      selectedIds = new SvelteSet([...selectedIds].filter((id) => !succeeded.has(id)));

      // Refresh affected libraries (counts + book lists)
      const affectedLibraries = new Set(
        items.filter((i) => succeeded.has(i.stagedBookId)).map((i) => i.libraryId)
      );
      librariesState.fetchAll();
      shelvesState.fetchAll();
      for (const libId of affectedLibraries) {
        booksState.invalidate(libId);
        booksState.fetchForLibrary(libId);
      }

      for (const id of succeeded) bookLibraryMap.delete(id);
      bookLibraryMap = new SvelteMap(bookLibraryMap);

      if (failed.length > 0) {
        errorMsg = failed
          .map((r) => {
            const book = stagedBooks.find((b) => b.id === r.stagedBookId);
            return `${book?.title ?? r.stagedBookId}: ${r.error}`;
          })
          .join('\n');
      }
    } catch (err: unknown) {
      errorMsg = err instanceof Error ? err.message : 'Import failed.';
    } finally {
      isImporting = false;
    }
  }

  // Bulk edit sheet
  let bulkSheetOpen = $state(false);
  let bulkAuthor = $state('');
  let isBulkSaving = $state(false);

  // Bulk add to library
  let selectedLibraryId = $state('');

  let selectedLibraryTitle = $derived(
    selectedLibraryId
      ? (librariesState.items.find((lib) => lib.id === selectedLibraryId)?.title ??
          'Select Library')
      : 'Select Library'
  );

  function openBulkEdit() {
    sheetOpen = false;
    bulkAuthor = '';
    selectedLibraryId = '';
    bulkSheetOpen = true;
  }

  async function saveBulkEdit() {
    isBulkSaving = true;
    try {
      if (bulkAuthor.trim()) {
        const items = [...selectedIds].map((id) => ({ id, author: bulkAuthor }));
        const res = await fetch('/api/bookdrop/staged', {
          method: 'PUT',
          headers: { Authorization: `Bearer ${getToken()}`, 'Content-Type': 'application/json' },
          body: JSON.stringify(items)
        });
        if (!res.ok) throw new Error('Bulk update failed');
        stagedBooks = await res.json();
      }
      if (selectedLibraryId) {
        // Assign the chosen library to all selected books then import
        for (const id of selectedIds) bookLibraryMap.set(id, selectedLibraryId);
        bookLibraryMap = new SvelteMap(bookLibraryMap);
        await handleAddAllToLibraries();
      }
      bulkSheetOpen = false;
      selectedIds = new SvelteSet();
    } catch {
      errorMsg = 'Bulk update failed.';
    } finally {
      isBulkSaving = false;
    }
  }

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

  async function loadStaged() {
    isLoading = true;
    errorMsg = null;
    try {
      const res = await fetch('/api/bookdrop/staged', {
        headers: { Authorization: `Bearer ${getToken()}` }
      });
      if (!res.ok) throw new Error('Failed to load staged books');
      stagedBooks = await res.json();
    } catch (err: unknown) {
      errorMsg = err instanceof Error ? err.message : 'Failed to load staged books.';
    } finally {
      isLoading = false;
    }
  }

  async function handleScan() {
    isScanning = true;
    errorMsg = null;
    try {
      const res = await fetch('/api/bookdrop/scan', {
        method: 'POST',
        headers: { Authorization: `Bearer ${getToken()}` }
      });
      if (!res.ok) throw new Error('Failed to scan bookdrop');
      stagedBooks = await res.json();
    } catch (err: unknown) {
      errorMsg = err instanceof Error ? err.message : 'An error occurred while scanning.';
    } finally {
      isScanning = false;
    }
  }

  function toggleAll() {
    if (allSelected) {
      selectedIds = new SvelteSet();
    } else {
      selectedIds = new SvelteSet(stagedBooks.map((b) => b.id));
    }
  }

  function toggleOne(id: string) {
    const next = new SvelteSet(selectedIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedIds = next;
  }

  function openEdit(book: StagedBook) {
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
      const res = await fetch(`/api/bookdrop/staged/${editingBook.id}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${getToken()}`,
          'Content-Type': 'application/json'
        },
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
      const updated: StagedBook = await res.json();
      stagedBooks = stagedBooks.map((b) => (b.id === updated.id ? updated : b));
      sheetOpen = false;
    } catch {
      errorMsg = 'Failed to save changes.';
    } finally {
      isSaving = false;
    }
  }

  async function handleDelete(id: string) {
    const res = await fetch(`/api/bookdrop/staged/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${getToken()}` }
    });
    if (res.ok) {
      stagedBooks = stagedBooks.filter((b) => b.id !== id);
      selectedIds.delete(id);
      selectedIds = new SvelteSet(selectedIds);
      if (editingBook?.id === id) sheetOpen = false;
    }
  }

  async function handleBulkDelete() {
    const ids = [...selectedIds];
    const results = await Promise.all(
      ids.map((id) =>
        fetch(`/api/bookdrop/staged/${id}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${getToken()}` }
        }).then((r) => (r.ok ? id : null))
      )
    );
    const deleted = new Set(results.filter((id) => id !== null));
    stagedBooks = stagedBooks.filter((b) => !deleted.has(b.id));
    selectedIds = new SvelteSet([...selectedIds].filter((id) => !deleted.has(id)));
    if (editingBook && deleted.has(editingBook.id)) sheetOpen = false;
  }

  function applyBulkLibrary(libId: string) {
    selectedLibraryId = libId;
    for (const id of selectedIds) {
      bookLibraryMap.set(id, libId);
    }
    bookLibraryMap = new SvelteMap(bookLibraryMap);
  }

  $effect(() => {
    loadStaged();
  });
</script>

<div class="flex flex-1 flex-col gap-4 p-4 pt-0">
  <div class="flex w-full items-center justify-end">
    <Button onclick={handleScan} disabled={isScanning}>
      {#if isScanning}
        <span>Scanning...</span>
        <Spinner />
      {:else}
        <span>Scan Bookdrop</span>
        <RotateCW />
      {/if}
    </Button>
  </div>
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 whitespace-pre-wrap text-destructive">
      {errorMsg}
    </div>
  {/if}

  <!-- Bulk action bar -->
  {#if selectedIds.size > 0}
    <div class="flex items-center gap-3 rounded-lg border bg-muted/50 px-4 py-2">
      <span class="text-sm text-muted-foreground">{selectedIds.size} selected</span>
      <Button size="sm" onclick={openBulkEdit}>Bulk Edit</Button>
      <div class="flex">
        <Select.Root type="single" bind:value={selectedLibraryId}>
          <Select.Trigger class="w- h-8">
            {selectedLibraryTitle}
          </Select.Trigger>
          <Select.Content>
            {#each librariesState.items as lib (lib.id)}
              <Select.Item value={lib.id}>{lib.title}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
        {#if selectedLibraryId}
          <Button variant="ghost" class="ml-2" onclick={() => applyBulkLibrary(selectedLibraryId)}>
            Apply Library
          </Button>
        {/if}
      </div>
      <Button size="sm" variant="destructive" onclick={handleBulkDelete}>Delete selected</Button>
      <Button size="sm" variant="ghost" onclick={() => (selectedIds = new SvelteSet())}>
        Clear
      </Button>
    </div>
  {/if}

  {#if isLoading}
    <div class="flex min-h-100 items-center justify-center">
      <p class="text-muted-foreground">Loading…</p>
    </div>
  {:else if stagedBooks.length > 0}
    <div class="rounded-lg border">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b bg-muted/50">
            <th class="w-10 px-4 py-3 text-left">
              <Checkbox checked={allSelected} onCheckedChange={toggleAll} />
            </th>
            <th class="w-12 px-2 py-3"></th>
            <th class="px-4 py-3 text-left font-medium">Title</th>
            <th class="px-4 py-3 text-left font-medium">Author</th>
            <th class="w-20 px-4 py-3 text-left font-medium">Type</th>
            <th class="max-w-xs px-4 py-3 text-left font-medium">File</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody>
          {#each stagedBooks as book (book.id)}
            <tr
              class="border-b transition-colors last:border-0 hover:bg-muted/30 {selectedIds.has(
                book.id
              )
                ? 'bg-muted/20'
                : ''}"
            >
              <td class="px-4 py-3">
                <Checkbox
                  checked={selectedIds.has(book.id)}
                  onCheckedChange={() => toggleOne(book.id)}
                />
              </td>
              <td class="px-2 py-2">
                {#if book.hasCover}
                  <img
                    src={coverUrl(book.id)}
                    alt=""
                    class="h-10 w-7 rounded object-cover shadow-sm"
                  />
                {:else}
                  <div class="h-10 w-7 rounded bg-muted"></div>
                {/if}
              </td>
              <td class="px-4 py-3 font-medium">{book.title}</td>
              <td class="px-4 py-3 text-muted-foreground">{book.author ?? '—'}</td>
              <td class="px-4 py-3 text-muted-foreground uppercase">{book.ext.slice(1)}</td>
              <td class="max-w-xs truncate px-4 py-3 text-muted-foreground" title={book.fileName}>
                {book.fileName}
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <Select.Root
                    type="single"
                    value={bookLibraryMap.get(book.id) ?? ''}
                    onValueChange={(v) => {
                      bookLibraryMap.set(book.id, v);
                      bookLibraryMap = new SvelteMap(bookLibraryMap);
                    }}
                  >
                    <Select.Trigger class="h-8 w-36 text-xs">
                      {getBookLibraryTitle(book.id)}
                    </Select.Trigger>
                    <Select.Content>
                      {#each librariesState.items as lib (lib.id)}
                        <Select.Item value={lib.id}>{lib.title}</Select.Item>
                      {/each}
                    </Select.Content>
                  </Select.Root>
                  <Button size="sm" variant="ghost" onclick={() => openEdit(book)}>Edit</Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    class="text-destructive hover:text-destructive"
                    onclick={() => handleDelete(book.id)}
                  >
                    ✕
                  </Button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
      {#if readyToImportCount > 0}
        <div class="flex justify-end border-t px-4 py-3">
          <Button onclick={handleAddAllToLibraries} disabled={isImporting}>
            {#if isImporting}
              Importing… <Spinner />
            {:else}
              Add {readyToImportCount} book{readyToImportCount === 1 ? '' : 's'} to {readyToImportCount ===
              1
                ? 'library'
                : 'libraries'}
            {/if}
          </Button>
        </div>
      {/if}
    </div>
  {:else}
    <div
      class="flex min-h-100 flex-1 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <p class="text-muted-foreground">Click "Scan Bookdrop" to find new files.</p>
    </div>
  {/if}
</div>

<!-- Bulk edit sheet (bottom) -->
<Sheet.Root
  bind:open={bulkSheetOpen}
  onOpenChange={(o) => {
    if (!o) {
      bulkAuthor = '';
      selectedLibraryId = '';
      isBulkSaving = false;
    }
  }}
>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="bottom" class="mx-auto max-w-2xl">
      <Sheet.Header>
        <Sheet.Title>Bulk Edit</Sheet.Title>
        <Sheet.Description class="text-xs text-muted-foreground">
          Editing {selectedIds.size} book{selectedIds.size === 1 ? '' : 's'}. Leave a field blank to
          keep existing values.
        </Sheet.Description>
      </Sheet.Header>
      <div class="grid grid-cols-2 gap-4 px-4 py-6">
        <div class="flex flex-col gap-1.5">
          <label for="bulk-author" class="text-sm font-medium">Author</label>
          <Input id="bulk-author" bind:value={bulkAuthor} placeholder="Set author for all…" />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium" for="bulk-library">Add to Library</label>
          <Select.Root name="bulk-library" type="single" bind:value={selectedLibraryId}>
            <Select.Trigger>
              {selectedLibraryTitle}
            </Select.Trigger>
            <Select.Content>
              {#each librariesState.items as lib (lib.id)}
                <Select.Item value={lib.id}>{lib.title}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
        </div>
      </div>
      <Sheet.Footer>
        <Sheet.Close>
          {#snippet child({ props })}
            <Button variant="outline" {...props}>Cancel</Button>
          {/snippet}
        </Sheet.Close>
        <Button
          onclick={saveBulkEdit}
          disabled={isBulkSaving || (!bulkAuthor.trim() && !selectedLibraryId)}
        >
          {isBulkSaving ? 'Saving…' : 'Apply'}
        </Button>
      </Sheet.Footer>
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>

<!-- Single book edit sheet -->
<Sheet.Root
  bind:open={sheetOpen}
  onOpenChange={(o) => {
    if (!o) {
      editingBook = null;
      editTitle = '';
      editAuthor = '';
      editSubject = '';
      editDescription = '';
      editPublisher = '';
      editContributor = '';
      editDate = '';
      editIdentifier = '';
      editLanguage = '';
      isSaving = false;
    }
  }}
>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96">
      {#if editingBook}
        <Sheet.Header>
          <Sheet.Title>Edit Book</Sheet.Title>
          <Sheet.Description class="truncate text-xs text-muted-foreground">
            {editingBook.fileName}
          </Sheet.Description>
        </Sheet.Header>
        <div class="flex flex-col gap-4 overflow-y-auto px-4 py-6">
          {#if editingBook.hasCover}
            <img
              src={coverUrl(editingBook.id)}
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
              <Input id="edit-date" bind:value={editDate} placeholder="YYYY or YYYY-MM-DD" />
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
