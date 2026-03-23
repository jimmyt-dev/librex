<script lang="ts">
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Separator } from '$lib/components/ui/separator';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { librariesState } from '$lib/api/libraries.svelte';
  import type { PageData } from './$types';
  import { SvelteSet } from 'svelte/reactivity';
  import * as Select from '$lib/components/ui/select';
  import RotateCW from '@lucide/svelte/icons/rotate-cw';

  interface StagedBook {
    id: string;
    title: string;
    author: string | null;
    fileName: string;
    ext: string;
    originalPath: string;
    userId: string;
  }

  let { data }: { data: PageData } = $props();

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
  let isSaving = $state(false);

  // Bulk add to library
  let selectedLibraryId = $state('');

  let selectedLibraryTitle = $derived(
    selectedLibraryId
      ? (librariesState.items.find((lib) => lib.id === selectedLibraryId)?.title ??
          'Select Library')
      : 'Select Library'
  );

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
        body: JSON.stringify({ title: editTitle, author: editAuthor || null })
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

  async function handleAddToLibrary() {
    if (!selectedLibraryId) return;
    // TODO: call import endpoint once available
    selectedLibraryId = '';
    selectedIds = new SvelteSet();
  }

  $effect(() => {
    loadStaged();
  });
</script>

<Sidebar.Provider>
  <AppSidebar user={data.user} />
  <Sidebar.Inset>
    <header
      class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
    >
      <div class="flex w-full items-center justify-between gap-2 px-4">
        <div class="flex items-center">
          <Sidebar.Trigger class="-ms-1" />
          <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb.Root>
            <Breadcrumb.List>
              <Breadcrumb.Item>
                <Breadcrumb.Page>Bookdrop</Breadcrumb.Page>
              </Breadcrumb.Item>
            </Breadcrumb.List>
          </Breadcrumb.Root>
        </div>
        <Button onclick={handleScan} disabled={isScanning}>
          <!-- {isScanning ? 'Scanning...' : 'Scan Bookdrop'} -->
          {#if isScanning}
            <span>Scanning...</span>
            <Spinner />
          {:else}
            <span>Scan Bookdrop</span>
            <RotateCW />
          {/if}
        </Button>
      </div>
    </header>

    <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
      {#if errorMsg}
        <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
      {/if}

      <!-- Bulk action bar -->
      {#if selectedIds.size > 0}
        <div class="flex items-center gap-3 rounded-lg border bg-muted/50 px-4 py-2">
          <span class="text-sm text-muted-foreground">{selectedIds.size} selected</span>

          <Select.Root type="single" bind:value={selectedLibraryId}>
            <Select.Trigger class="w-[180px]">
              {selectedLibraryTitle}
            </Select.Trigger>
            <Select.Content>
              {#each librariesState.items as lib (lib.id)}
                <Select.Item value={lib.id}>{lib.title}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>

          <Button onclick={handleAddToLibrary} disabled={!selectedLibraryId}>Add to Library</Button>
          <Button variant="destructive" onclick={handleBulkDelete}>Delete selected</Button>
          <Button variant="ghost" onclick={() => (selectedIds = new SvelteSet())}>Clear</Button>
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
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onchange={toggleAll}
                    class="rounded"
                  />
                </th>
                <th class="px-4 py-3 text-left font-medium">Title</th>
                <th class="px-4 py-3 text-left font-medium">Author</th>
                <th class="w-20 px-4 py-3 text-left font-medium">Type</th>
                <th class="max-w-xs px-4 py-3 text-left font-medium">File</th>
                <th class="w-24 px-4 py-3"></th>
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
                    <input
                      type="checkbox"
                      checked={selectedIds.has(book.id)}
                      onchange={() => toggleOne(book.id)}
                      class="rounded"
                    />
                  </td>
                  <td class="px-4 py-3 font-medium">{book.title}</td>
                  <td class="px-4 py-3 text-muted-foreground">{book.author ?? '—'}</td>
                  <td class="px-4 py-3 text-muted-foreground uppercase">{book.ext.slice(1)}</td>
                  <td
                    class="max-w-xs truncate px-4 py-3 text-muted-foreground"
                    title={book.fileName}
                  >
                    {book.fileName}
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex justify-end gap-1">
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
        </div>
      {:else}
        <div
          class="flex min-h-[400px] flex-1 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
        >
          <p class="text-muted-foreground">Click "Scan Bookdrop" to find new files.</p>
        </div>
      {/if}
    </div>
  </Sidebar.Inset>
</Sidebar.Provider>

<!-- Single book edit sheet -->
<Sheet.Root bind:open={sheetOpen}>
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
        <div class="flex flex-col gap-4 px-4 py-6">
          <div class="flex flex-col gap-1.5">
            <label for="edit-title" class="text-sm font-medium">Title</label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <label for="edit-author" class="text-sm font-medium">Author</label>
            <Input id="edit-author" bind:value={editAuthor} placeholder="Unknown" />
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
