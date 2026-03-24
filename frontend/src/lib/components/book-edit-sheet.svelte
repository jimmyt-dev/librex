<script lang="ts">
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import * as Sheet from '$lib/components/ui/sheet';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

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
  let errorMsg = $state<string | null>(null);

  $effect(() => {
    const book = bookEditState.book;
    if (!book) return;
    editTitle = book.title;
    editAuthor = book.author ?? '';
    editSubject = book.subject ?? '';
    editDescription = book.description ?? '';
    editPublisher = book.publisher ?? '';
    editContributor = book.contributor ?? '';
    editDate = book.date ?? '';
    editIdentifier = book.identifier ?? '';
    editLanguage = book.language ?? '';
    errorMsg = null;
  });

  async function saveEdit() {
    if (!bookEditState.book) return;
    isSaving = true;
    errorMsg = null;
    try {
      const res = await fetch(`/api/books/${bookEditState.book.id}`, {
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
      const updated = await res.json();
      booksState.upsert(updated);
      bookEditState.close();
    } catch {
      errorMsg = 'Failed to save changes.';
    } finally {
      isSaving = false;
    }
  }
</script>

<Sheet.Root bind:open={bookEditState.open}>
  <Sheet.Portal>
    <Sheet.Overlay />
    <Sheet.Content side="right" class="w-96 overflow-y-auto">
      {#if bookEditState.book}
        <Sheet.Header>
          <Sheet.Title>Edit Metadata</Sheet.Title>
          <Sheet.Description class="truncate text-xs text-muted-foreground">
            {bookEditState.book.filePath.split('/').pop()}
          </Sheet.Description>
        </Sheet.Header>
        <div class="flex flex-col gap-4 overflow-y-auto px-4 py-6">
          {#if bookEditState.book.cover}
            <img
              src={bookEditState.book.cover}
              alt="Cover"
              class="mx-auto h-48 w-auto rounded object-contain shadow"
            />
          {/if}
          {#if errorMsg}
            <p class="text-sm text-destructive">{errorMsg}</p>
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
