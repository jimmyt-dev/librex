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
    fetchTagSuggestions,
    fetchSeriesSuggestions
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
  import { Textarea } from '$lib/components/ui/textarea';
  import { toast } from 'svelte-sonner';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';

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

  function normalizeDate(d: string | null | undefined): string {
    if (!d) return '';
    if (/^\d{4}$/.test(d)) return `${d}-01-01`;
    if (/^\d{4}-\d{2}$/.test(d)) return `${d}-01`;
    return d;
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
  let deleteDialogOpen = $state(false);
  let deleteTargetId = $state<string | null>(null);
  let deleteTargetTitle = $state('');
  let deleteFile = $state(false);
  let deleting = $state(false);
  let isBulkDelete = $state(false);
  let seriesSuggestions = $state<string[]>([]);
  let showSeriesDropdown = $state(false);
  let seriesHighlightIndex = $state(-1);

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
    editDate = normalizeDate(book.date);
    editIdentifier = book.identifier ?? '';
    editLanguage = book.language ?? '';
    editSeriesName = book.seriesName ?? '';
    editSeriesNumber = book.seriesNumber?.toString() ?? '';
    editSeriesTotal = book.seriesTotal?.toString() ?? '';
    editPageCount = book.pageCount?.toString() ?? '';
    editRating = book.rating?.toString() ?? '';
    seriesSuggestions = [];
    showSeriesDropdown = false;
    seriesHighlightIndex = -1;
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

  function handleDelete(book: StagedBook) {
    deleteTargetId = book.id;
    deleteTargetTitle = book.title;
    deleteFile = false;
    deleteDialogOpen = true;
  }

  async function confirmDelete() {
    deleting = true;
    try {
      if (isBulkDelete) {
        const ids = [...selectedIds];
        const suffix = deleteFile ? '?deleteFile=true' : '';
        const results = await Promise.all(
          ids.map((id) =>
            apiFetch(`/api/bookdrop/staged/${id}${suffix}`, { method: 'DELETE' })
              .then(() => id)
              .catch(() => null)
          )
        );
        const deleted = new Set(results.filter((id) => id !== null));
        stagedBooks = stagedBooks.filter((b) => !deleted.has(b.id));
        selectedIds = new SvelteSet([...selectedIds].filter((id) => !deleted.has(id)));
        if (editingBook && deleted.has(editingBook.id)) sheetOpen = false;
      } else {
        if (!deleteTargetId) return;
        const url = `/api/bookdrop/staged/${deleteTargetId}${deleteFile ? '?deleteFile=true' : ''}`;
        await apiFetch(url, { method: 'DELETE' });
        stagedBooks = stagedBooks.filter((b) => b.id !== deleteTargetId);
        selectedIds.delete(deleteTargetId);
        selectedIds = new SvelteSet(selectedIds);
        if (editingBook?.id === deleteTargetId) sheetOpen = false;
      }
      deleteDialogOpen = false;
    } catch {
      toast.error('Failed to delete.');
    } finally {
      deleting = false;
    }
  }

  function handleBulkDelete() {
    isBulkDelete = true;
    deleteTargetId = null;
    deleteTargetTitle = '';
    deleteFile = false;
    deleteDialogOpen = true;
  }

  function applyBulkLibrary(libId: string) {
    selectedLibraryId = libId;
    for (const id of selectedIds) {
      bookLibraryMap.set(id, libId);
    }
    bookLibraryMap = new SvelteMap(bookLibraryMap);
  }

  let maxFileSizeMB = $state(100);

  $effect(() => {
    loadStaged();
    apiFetch('/api/settings')
      .then((s: { maxUploadSizeMb?: number }) => {
        if (s.maxUploadSizeMb) maxFileSizeMB = s.maxUploadSizeMb;
      })
      .catch(() => {});
  });
</script>

<div class="page-content gap-4">
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
        {maxFileSizeMB}
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
    <div class="flex flex-wrap items-center gap-2 rounded-lg border bg-muted/50 px-4 py-2">
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
          <Button size="sm" variant="ghost" onclick={() => applyBulkLibrary(selectedLibraryId)}>
            Apply
          </Button>
        {/if}
      </div>
      <Button size="sm" variant="destructive" onclick={handleBulkDelete}>Delete</Button>
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
    <!-- Mobile card list -->
    <div class="flex flex-col gap-2 md:hidden">
      {#each stagedBooks as book (book.id)}
        <div
          class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2 {selectedIds.has(
            book.id
          )
            ? 'ring-1 ring-primary'
            : ''}"
        >
          <Checkbox checked={selectedIds.has(book.id)} onCheckedChange={() => toggleOne(book.id)} />
          {#if book.hasCover}
            <img
              src={coverUrl(book.id)}
              alt=""
              class="h-12 w-8 shrink-0 rounded object-cover shadow-sm"
            />
          {:else}
            <div class="h-12 w-8 shrink-0 rounded bg-muted"></div>
          {/if}
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium">{book.title}</p>
            <p class="truncate text-xs text-muted-foreground">
              {book.author ?? '—'} · {book.ext.slice(1).toUpperCase()}
            </p>
            <Select.Root
              type="single"
              value={bookLibraryMap.get(book.id) ?? ''}
              onValueChange={(v) => {
                bookLibraryMap.set(book.id, v);
                bookLibraryMap = new SvelteMap(bookLibraryMap);
              }}
            >
              <Select.Trigger class="mt-1 h-7 text-xs">
                {getBookLibraryTitle(book.id)}
              </Select.Trigger>
              <Select.Content>
                {#each librariesState.items as lib (lib.id)}
                  <Select.Item value={lib.id}>{lib.title}</Select.Item>
                {/each}
              </Select.Content>
            </Select.Root>
          </div>
          <div class="flex shrink-0 flex-col gap-1">
            <Button size="sm" variant="ghost" onclick={() => openEdit(book)}>Edit</Button>
            <Button
              size="sm"
              variant="ghost"
              class="text-destructive hover:text-destructive"
              onclick={() => handleDelete(book)}>✕</Button
            >
          </div>
        </div>
      {/each}
      {#if readyToImportCount > 0}
        <div class="flex justify-end pt-1">
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

    <!-- Desktop table -->
    <div class="hidden rounded-lg border md:block">
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
                    onclick={() => handleDelete(book)}
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
    <!-- /Desktop table -->
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
            <Textarea
              id="edit-description"
              bind:value={editDescription}
              placeholder="Description"
            />
          </div>
          <div class="flex flex-col gap-1.5">
            <Label for="edit-publisher" class="text-sm font-medium">Publisher</Label>
            <Input id="edit-publisher" bind:value={editPublisher} />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1.5">
              <Label for="edit-date" class="text-sm font-medium">Published Date</Label>
              <Input id="edit-date" type="date" bind:value={editDate} />
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
              <div class="relative">
                <Input
                  bind:value={editSeriesName}
                  placeholder="Series name"
                  oninput={async () => {
                    seriesHighlightIndex = -1;
                    if (editSeriesName.trim().length < 1) {
                      seriesSuggestions = [];
                      showSeriesDropdown = false;
                      return;
                    }
                    seriesSuggestions = await fetchSeriesSuggestions(editSeriesName.trim());
                    showSeriesDropdown = seriesSuggestions.length > 0;
                  }}
                  onkeydown={(e) => {
                    if (e.key === 'ArrowDown') {
                      e.preventDefault();
                      if (seriesSuggestions.length > 0) {
                        showSeriesDropdown = true;
                        seriesHighlightIndex = Math.min(
                          seriesHighlightIndex + 1,
                          seriesSuggestions.length - 1
                        );
                      }
                    } else if (e.key === 'ArrowUp') {
                      e.preventDefault();
                      seriesHighlightIndex = Math.max(seriesHighlightIndex - 1, -1);
                    } else if (e.key === 'Enter' && seriesHighlightIndex >= 0) {
                      e.preventDefault();
                      editSeriesName = seriesSuggestions[seriesHighlightIndex];
                      showSeriesDropdown = false;
                      seriesHighlightIndex = -1;
                    } else if (e.key === 'Escape') {
                      showSeriesDropdown = false;
                      seriesHighlightIndex = -1;
                    }
                  }}
                  onfocus={() => {
                    if (seriesSuggestions.length > 0) showSeriesDropdown = true;
                  }}
                  onblur={() => setTimeout(() => (showSeriesDropdown = false), 150)}
                />
                {#if showSeriesDropdown}
                  <div
                    class="absolute top-full right-0 left-0 z-50 mt-1 max-h-32 overflow-y-auto rounded-lg border bg-popover shadow-md"
                  >
                    {#each seriesSuggestions as s, i (i)}
                      <button
                        type="button"
                        class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i ===
                        seriesHighlightIndex
                          ? 'bg-accent'
                          : ''}"
                        onmousedown={() => {
                          editSeriesName = s;
                          showSeriesDropdown = false;
                          seriesHighlightIndex = -1;
                        }}>{s}</button
                      >
                    {/each}
                  </div>
                {/if}
              </div>
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
          {#if dirty}
            <Button variant="outline" onclick={revertEdit}>
              <RotateCcwIcon class="size-4" />
              Revert
            </Button>
          {/if}
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

<AlertDialog.Root
  open={deleteDialogOpen}
  onOpenChange={(o) => {
    if (!o) {
      deleteDialogOpen = false;
      deleteFile = false;
      isBulkDelete = false;
    }
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>
        {isBulkDelete ? `Delete ${selectedIds.size} books?` : `Delete "${deleteTargetTitle}"?`}
      </AlertDialog.Title>
      <AlertDialog.Description>
        This will remove the {isBulkDelete ? 'books' : 'book'} from the bookdrop queue. This action cannot
        be undone.
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
