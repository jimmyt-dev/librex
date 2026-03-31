<script lang="ts">
  import * as Sheet from '$lib/components/ui/sheet';
  import { headerState } from '$lib/state/header.svelte';
  headerState.title = 'Bookdrop';
  headerState.subtitle = null;
  headerState.counts = [];
  import { apiFetch } from '$lib/api/client';
  import {
    fetchAuthorSuggestions,
    fetchGenreSuggestions,
    fetchTagSuggestions
  } from '$lib/api/suggestions';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { SvelteMap, SvelteSet } from 'svelte/reactivity';
  import * as Select from '$lib/components/ui/select';
  import RotateCW from '@lucide/svelte/icons/rotate-cw';
  import FileUpload from '$lib/components/file-upload.svelte';
  import UploadIcon from '@lucide/svelte/icons/upload';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import { Spinner } from '$lib/components/ui/spinner';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import TagInput from '$lib/components/tag-input.svelte';
  import ArrayField from '$lib/components/array-field.svelte';
  import StarRating from '$lib/components/star-rating.svelte';
  import { Label } from '$lib/components/ui/label';
  import { toast } from 'svelte-sonner';

  interface StagedBook {
    id: string;
    title: string;
    subtitle: string | null;
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
    seriesName: string | null;
    seriesNumber: number | null;
    seriesTotal: number | null;
    pageCount: number | null;
    rating: number | null;
    tags: string | null;
    hasCover: boolean;
    fileName: string;
    ext: string;
    originalPath: string;
    userId: string;
  }

  function coverUrl(id: string) {
    return `/api/bookdrop/staged/${id}/cover`;
  }

  function subjectToTags(subject: string | null): string[] {
    if (!subject) return [];
    return subject
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
  }

  function tagsToSubject(tags: string[]): string {
    return tags.join(', ');
  }

  let stagedBooks = $state<StagedBook[]>([]);
  let isScanning = $state(false);
  let uploadOpen = $state(false);
  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

  // Selection
  let selectedIds = $state<Set<string>>(new Set());
  let allSelected = $derived(stagedBooks.length > 0 && selectedIds.size === stagedBooks.length);

  // Single-book sheet edit
  let sheetOpen = $state(false);
  let editingBook = $state<StagedBook | null>(null);
  let editTitle = $state('');
  let editSubtitle = $state('');
  let editAuthors = $state<string[]>([]);
  let editGenres = $state<string[]>([]);
  let editTags = $state<string[]>([]);
  let editDescription = $state('');
  let editPublisher = $state('');
  let editDate = $state('');
  let editIdentifier = $state('');
  let editLanguage = $state('');
  let editSeriesName = $state('');
  let editSeriesNumber = $state('');
  let editSeriesTotal = $state('');
  let editPageCount = $state('');
  let editRating = $state('');
  let isSaving = $state(false);

  let dirty = $derived.by(() => {
    const b = editingBook;
    if (!b) return false;
    return (
      editTitle !== b.title ||
      editSubtitle !== (b.subtitle ?? '') ||
      editDescription !== (b.description ?? '') ||
      editPublisher !== (b.publisher ?? '') ||
      editDate !== (b.date ?? '') ||
      editIdentifier !== (b.identifier ?? '') ||
      editLanguage !== (b.language ?? '') ||
      editPageCount !== (b.pageCount?.toString() ?? '') ||
      editSeriesName !== (b.seriesName ?? '') ||
      editSeriesNumber !== (b.seriesNumber?.toString() ?? '') ||
      editSeriesTotal !== (b.seriesTotal?.toString() ?? '') ||
      editRating !== (b.rating?.toString() ?? '') ||
      JSON.stringify(editAuthors) !== JSON.stringify(subjectToTags(b.author)) ||
      JSON.stringify(editGenres) !== JSON.stringify(subjectToTags(b.subject)) ||
      JSON.stringify(editTags) !== JSON.stringify(subjectToTags(b.tags))
    );
  });

  function revertEdit() {
    if (!editingBook) return;
    openEdit(editingBook);
  }

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
      const results: { stagedBookId: string; error?: string }[] = await apiFetch(
        '/api/bookdrop/import',
        { method: 'POST', body: JSON.stringify(items) }
      );

      const succeeded = new Set(results.filter((r) => !r.error).map((r) => r.stagedBookId));
      const failed = results.filter((r) => r.error);

      stagedBooks = stagedBooks.filter((b) => !succeeded.has(b.id));
      selectedIds = new SvelteSet([...selectedIds].filter((id) => !succeeded.has(id)));

      const affectedLibraries = new Set(
        items.filter((i) => succeeded.has(i.stagedBookId)).map((i) => i.libraryId)
      );
      librariesState.fetchAll();
      shelvesState.fetchAll();
      booksState.fetchAll();
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
  let bulkSeriesName = $state('');
  let bulkPublisher = $state('');
  let bulkLanguage = $state('');
  let bulkSeriesTotal = $state('');
  let bulkAuthors = $state<string[]>([]);
  let bulkAuthorsMode = $state<'replace' | 'merge'>('merge');
  let bulkGenres = $state<string[]>([]);
  let bulkGenresMode = $state<'replace' | 'merge'>('merge');
  let bulkTags = $state<string[]>([]);
  let bulkTagsMode = $state<'replace' | 'merge'>('merge');
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
    bulkSeriesName = '';
    bulkPublisher = '';
    bulkLanguage = '';
    bulkSeriesTotal = '';
    bulkAuthors = [];
    bulkAuthorsMode = 'merge';
    bulkGenres = [];
    bulkGenresMode = 'merge';
    bulkTags = [];
    bulkTagsMode = 'merge';
    selectedLibraryId = '';
    bulkSheetOpen = true;
  }

  function bulkHasChanges() {
    return (
      bulkSeriesName.trim() !== '' ||
      bulkPublisher.trim() !== '' ||
      bulkLanguage.trim() !== '' ||
      bulkSeriesTotal !== '' ||
      bulkAuthors.length > 0 ||
      bulkGenres.length > 0 ||
      bulkTags.length > 0 ||
      !!selectedLibraryId
    );
  }

  async function saveBulkEdit() {
    isBulkSaving = true;
    try {
      const hasFieldUpdate =
        bulkSeriesName.trim() ||
        bulkPublisher.trim() ||
        bulkLanguage.trim() ||
        bulkSeriesTotal !== '' ||
        bulkAuthors.length > 0 ||
        bulkGenres.length > 0 ||
        bulkTags.length > 0;

      if (hasFieldUpdate) {
        const payload: Record<string, unknown> = { ids: [...selectedIds] };
        if (bulkSeriesName.trim()) payload.seriesName = bulkSeriesName.trim();
        if (bulkPublisher.trim()) payload.publisher = bulkPublisher.trim();
        if (bulkLanguage.trim()) payload.language = bulkLanguage.trim();
        if (bulkSeriesTotal !== '') payload.seriesTotal = parseInt(bulkSeriesTotal);
        if (bulkAuthors.length > 0) {
          payload.authors = bulkAuthors;
          payload.authorsMode = bulkAuthorsMode;
        }
        if (bulkGenres.length > 0) {
          payload.genres = bulkGenres;
          payload.genresMode = bulkGenresMode;
        }
        if (bulkTags.length > 0) {
          payload.tags = bulkTags;
          payload.tagsMode = bulkTagsMode;
        }

        stagedBooks = await apiFetch('/api/bookdrop/staged', {
          method: 'PUT',
          body: JSON.stringify(payload)
        });
      }

      if (selectedLibraryId) {
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

  async function loadStaged() {
    isLoading = true;
    errorMsg = null;
    try {
      stagedBooks = await apiFetch('/api/bookdrop/staged');
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
      stagedBooks = await apiFetch('/api/bookdrop/scan', { method: 'POST' });
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
    editSubtitle = book.subtitle ?? '';
    editAuthors = subjectToTags(book.author);
    editGenres = subjectToTags(book.subject);
    editTags = subjectToTags(book.tags);
    editDescription = book.description ?? '';
    editPublisher = book.publisher ?? '';
    editDate = book.date ?? '';
    editIdentifier = book.identifier ?? '';
    editLanguage = book.language ?? '';
    editSeriesName = book.seriesName ?? '';
    editSeriesNumber = book.seriesNumber?.toString() ?? '';
    editSeriesTotal = book.seriesTotal?.toString() ?? '';
    editPageCount = book.pageCount?.toString() ?? '';
    editRating = book.rating?.toString() ?? '';
    sheetOpen = true;
  }

  async function saveEdit() {
    if (!editingBook) return;
    isSaving = true;
    try {
      const updated: StagedBook = await apiFetch(`/api/bookdrop/staged/${editingBook.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          title: editTitle,
          subtitle: editSubtitle || null,
          author: editAuthors.length > 0 ? tagsToSubject(editAuthors) : null,
          subject: editGenres.length > 0 ? tagsToSubject(editGenres) : null,
          tags: editTags.length > 0 ? tagsToSubject(editTags) : null,
          description: editDescription || null,
          publisher: editPublisher || null,
          date: editDate || null,
          identifier: editIdentifier || null,
          language: editLanguage || null,
          seriesName: editSeriesName || null,
          seriesNumber: editSeriesNumber ? parseFloat(editSeriesNumber) : null,
          seriesTotal: editSeriesTotal ? parseInt(editSeriesTotal) : null,
          pageCount: editPageCount ? parseInt(editPageCount) : null,
          rating: editRating ? parseInt(editRating) : null
        })
      });
      stagedBooks = stagedBooks.map((b) => (b.id === updated.id ? updated : b));
      sheetOpen = false;
    } catch {
      errorMsg = 'Failed to save changes.';
    } finally {
      isSaving = false;
    }
  }

  async function handleDelete(id: string) {
    try {
      await apiFetch(`/api/bookdrop/staged/${id}`, { method: 'DELETE' });
      stagedBooks = stagedBooks.filter((b) => b.id !== id);
      selectedIds.delete(id);
      selectedIds = new SvelteSet(selectedIds);
      if (editingBook?.id === id) sheetOpen = false;
    } catch {
      // silently ignore
    }
  }

  async function handleBulkDelete() {
    const ids = [...selectedIds];
    const results = await Promise.all(
      ids.map((id) =>
        apiFetch(`/api/bookdrop/staged/${id}`, { method: 'DELETE' })
          .then(() => id)
          .catch(() => null)
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

  // Page-level drag and drop
  let pageDragOver = $state(false);
  let dragCounter = 0; // track enters/leaves for nested elements

  async function handlePageDrop(e: DragEvent) {
    e.preventDefault();
    pageDragOver = false;
    dragCounter = 0;
    const files = e.dataTransfer?.files;
    if (!files || files.length === 0) return;

    const formData = new FormData();
    let added = 0;
    for (const f of Array.from(files)) {
      const ext = '.' + f.name.split('.').pop()?.toLowerCase();
      if (['.epub', '.pdf', '.mobi', '.azw3', '.cbz', '.cbr'].includes(ext)) {
        formData.append('files', f);
        added++;
      }
    }
    if (added === 0) return;
    try {
      const result: StagedBook[] = await apiFetch('/api/bookdrop/upload', {
        method: 'POST',
        body: formData
      });
      stagedBooks = result;
      toast.success(`${added} file${added !== 1 ? 's' : ''} uploaded to bookdrop.`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Upload failed.');
    }
  }
</script>

<div
  role="region"
  aria-label="Bookdrop upload area"
  class="relative flex flex-1 flex-col gap-4 p-4 pt-0"
  ondragenter={(e) => {
    e.preventDefault();
    dragCounter++;
    pageDragOver = true;
  }}
  ondragleave={() => {
    dragCounter--;
    if (dragCounter <= 0) {
      pageDragOver = false;
      dragCounter = 0;
    }
  }}
  ondragover={(e) => e.preventDefault()}
  ondrop={handlePageDrop}
>
  {#if pageDragOver}
    <div
      class="pointer-events-none absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed border-primary bg-primary/5"
    >
      <UploadIcon class="size-10 text-primary" />
      <p class="text-base font-medium text-primary">Drop to upload to bookdrop</p>
    </div>
  {/if}

  <div class="flex w-full items-center justify-end gap-2">
    <Button variant="outline" onclick={() => (uploadOpen = !uploadOpen)}>
      <UploadIcon class="size-4" />
      Upload Files
    </Button>
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

  {#if uploadOpen}
    <div class="rounded-lg border p-4">
      <FileUpload
        onUploaded={(result) => {
          stagedBooks = result as typeof stagedBooks;
          uploadOpen = false;
        }}
      />
    </div>
  {/if}
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
          <Select.Trigger class="h-8">
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

<!-- Bulk edit sheet -->
<Sheet.Root
  bind:open={bulkSheetOpen}
  onOpenChange={(o) => {
    if (!o) {
      bulkSeriesName = '';
      bulkPublisher = '';
      bulkLanguage = '';
      bulkSeriesTotal = '';
      bulkAuthors = [];
      bulkAuthorsMode = 'merge';
      bulkGenres = [];
      bulkGenresMode = 'merge';
      bulkTags = [];
      bulkTagsMode = 'merge';
      selectedLibraryId = '';
      isBulkSaving = false;
    }
  }}
>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      <Sheet.Header>
        <Sheet.Title>Bulk Edit</Sheet.Title>
        <Sheet.Description class="text-xs text-muted-foreground">
          Editing {selectedIds.size} book{selectedIds.size === 1 ? '' : 's'}. Leave a field blank to
          keep existing values.
        </Sheet.Description>
      </Sheet.Header>
      <div class="flex flex-col gap-4 px-4 py-6">
        <!-- Text fields -->
        <p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">Text Fields</p>
        <div class="flex flex-col gap-1.5">
          <Label for="bulk-series" class="text-sm font-medium">Series Name</Label>
          <Input id="bulk-series" bind:value={bulkSeriesName} placeholder="Leave empty to skip" />
        </div>
        <div class="flex flex-col gap-1.5">
          <Label for="bulk-publisher" class="text-sm font-medium">Publisher</Label>
          <Input id="bulk-publisher" bind:value={bulkPublisher} placeholder="Leave empty to skip" />
        </div>
        <div class="flex flex-col gap-1.5">
          <Label for="bulk-language" class="text-sm font-medium">Language</Label>
          <Input id="bulk-language" bind:value={bulkLanguage} placeholder="e.g. en" />
        </div>

        <!-- Number fields -->
        <p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
          Number Fields
        </p>
        <div class="flex flex-col gap-1.5">
          <Label for="bulk-series-total" class="text-sm font-medium">Series Total</Label>
          <Input
            id="bulk-series-total"
            type="number"
            bind:value={bulkSeriesTotal}
            placeholder="—"
          />
        </div>

        <div class="h-px bg-border"></div>

        <!-- Array fields -->
        <p class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
          Array Fields
        </p>

        <ArrayField
          label="Authors"
          bind:values={bulkAuthors}
          bind:mode={bulkAuthorsMode}
          placeholder="Type and press Enter to add each item."
          fetchSuggestions={fetchAuthorSuggestions}
        />
        <ArrayField
          label="Genres"
          bind:values={bulkGenres}
          bind:mode={bulkGenresMode}
          placeholder="Type and press Enter to add each item."
          fetchSuggestions={fetchGenreSuggestions}
        />
        <ArrayField
          label="Tags"
          bind:values={bulkTags}
          bind:mode={bulkTagsMode}
          placeholder="Type and press Enter to add each item."
          fetchSuggestions={fetchTagSuggestions}
        />

        <div class="h-px bg-border"></div>

        <div class="flex flex-col gap-1.5">
          <Label class="text-sm font-medium" for="bulk-library">Add to Library &amp; Import</Label>
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
          <p class="text-xs text-muted-foreground">
            Selecting a library will import all selected books immediately after saving.
          </p>
        </div>
      </div>
      <Sheet.Footer>
        <Sheet.Close>
          {#snippet child({ props })}
            <Button variant="outline" {...props}>Cancel</Button>
          {/snippet}
        </Sheet.Close>
        <Button onclick={saveBulkEdit} disabled={isBulkSaving || !bulkHasChanges()}>
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
      editSubtitle = '';
      editAuthors = [];
      editGenres = [];
      editTags = [];
      editDescription = '';
      editPublisher = '';
      editDate = '';
      editIdentifier = '';
      editLanguage = '';
      editSeriesName = '';
      editSeriesNumber = '';
      editSeriesTotal = '';
      editPageCount = '';
      editRating = '';
      isSaving = false;
    }
  }}
>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
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
            <Label for="edit-title" class="text-sm font-medium">Title</Label>
            <Input id="edit-title" bind:value={editTitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-subtitle" class="text-sm font-medium">Subtitle</Label>
            <Input id="edit-subtitle" bind:value={editSubtitle} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label class="text-sm font-medium">Authors</Label>
            <TagInput
              bind:values={editAuthors}
              placeholder="Add author…"
              fetchSuggestions={fetchAuthorSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-description" class="text-sm font-medium">Description</Label>
            <Input id="edit-description" bind:value={editDescription} placeholder="Synopsis" />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-publisher" class="text-sm font-medium">Publisher</Label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <Label for="edit-date" class="text-sm font-medium">Published Date</Label>
              <Input id="edit-date" bind:value={editDate} placeholder="YYYY" />
            </div>
            <div class="flex flex-col gap-1.5">
              <Label for="edit-language" class="text-sm font-medium">Language</Label>
              <Input id="edit-language" bind:value={editLanguage} placeholder="en" />
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-identifier" class="text-sm font-medium">ISBN</Label>
            <Input id="edit-identifier" bind:value={editIdentifier} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-page-count" class="text-sm font-medium">Page Count</Label>
            <Input id="edit-page-count" type="number" bind:value={editPageCount} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label class="text-sm font-medium">Series</Label>
            <div class="grid grid-cols-[1fr_4rem] gap-2">
              <Input bind:value={editSeriesName} placeholder="Series name" />
              <Input type="number" bind:value={editSeriesNumber} placeholder="#" />
            </div>
            <div class="mt-1 grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <Label class="text-xs text-muted-foreground">Total Books</Label>
                <Input type="number" bind:value={editSeriesTotal} placeholder="Total" />
              </div>
            </div>
          </div>
          <div class="flex flex-col gap-1.5">
            <Label class="text-sm font-medium">Rating</Label>
            <StarRating bind:value={editRating} />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label class="text-sm font-medium">Genres</Label>
            <TagInput
              bind:values={editGenres}
              placeholder="Add genre…"
              fetchSuggestions={fetchGenreSuggestions}
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label class="text-sm font-medium">Tags</Label>
            <TagInput
              bind:values={editTags}
              placeholder="Add tag…"
              fetchSuggestions={fetchTagSuggestions}
            />
          </div>
        </div>
        <Sheet.Footer>
          <Sheet.Close>
            {#snippet child({ props })}
              <Button variant="outline" {...props}>Cancel</Button>
            {/snippet}
          </Sheet.Close>
          {#if dirty}
            <Button variant="ghost" size="icon" onclick={revertEdit} title="Revert changes">
              <RotateCcwIcon class="size-4" />
            </Button>
          {/if}
          <Button onclick={saveEdit} disabled={isSaving || !editTitle.trim()}>
            {isSaving ? 'Saving…' : 'Save'}
          </Button>
        </Sheet.Footer>
      {/if}
    </Sheet.Content>
  </Sheet.Portal>
</Sheet.Root>
