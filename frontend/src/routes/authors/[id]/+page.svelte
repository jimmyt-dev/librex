<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { authorsState, type Author } from '$lib/api/authors.svelte';
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';
  import BookView from '$lib/components/book-view.svelte';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { toast } from 'svelte-sonner';
  import UserIcon from '@lucide/svelte/icons/user';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import BookOpenIcon from '@lucide/svelte/icons/book-open';

  let authorId = $derived(page.params.id ?? '');

  let author = $state<Author | null>(null);
  let books = $state<Book[]>([]);
  let isLoading = $state(true);
  let errorMsg = $state<string | null>(null);

  let renameOpen = $state(false);
  let renameName = $state('');
  let renaming = $state(false);
  let deleteOpen = $state(false);
  let deleting = $state(false);

  $effect(() => {
    headerState.title = author?.name ?? 'Author';
    headerState.subtitle = null;
    headerState.counts = isLoading ? [] : [{ label: 'books', value: books.length }];
  });

  $effect(() => {
    const id = authorId;
    isLoading = true;
    errorMsg = null;
    authorsState
      .fetchOne(id)
      .then((a) => {
        author = a;
        renameName = a.name;
        return authorsState.fetchBooksForAuthor(id);
      })
      .then((b) => {
        books = b;
      })
      .catch((e: unknown) => {
        errorMsg = e instanceof Error ? e.message : 'Failed to load author.';
      })
      .finally(() => {
        isLoading = false;
      });
  });

  async function confirmRename() {
    if (!author || !renameName.trim()) return;
    renaming = true;
    try {
      await authorsState.update(author.id, renameName.trim());
      author = { ...author, name: renameName.trim() };
      toast.success(`Renamed to "${renameName.trim()}".`);
      renameOpen = false;
    } catch {
      toast.error('Failed to rename author.');
    } finally {
      renaming = false;
    }
  }

  async function confirmDelete() {
    if (!author) return;
    deleting = true;
    try {
      await authorsState.delete(author.id);
      toast.success(`"${author.name}" deleted.`);
      goto('/authors');
    } catch {
      toast.error('Failed to delete author.');
      deleting = false;
    }
  }
</script>

<div class="flex flex-1 flex-col gap-6 p-4 pt-0">
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
  {/if}

  {#if isLoading}
    <div class="flex items-center gap-4">
      <Skeleton class="size-20 rounded-full" />
      <div class="flex flex-col gap-2">
        <Skeleton class="h-7 w-48" />
        <Skeleton class="h-4 w-24" />
      </div>
    </div>
  {:else if author}
    <!-- Author header -->
    <div class="flex items-center gap-4">
      <div
        class="flex size-20 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground"
      >
        <UserIcon class="size-9" />
      </div>
      <div class="flex min-w-0 flex-1 flex-col gap-1">
        <h1 class="truncate text-2xl font-bold">{author.name}</h1>
        <p class="flex items-center gap-1.5 text-sm text-muted-foreground">
          <BookOpenIcon class="size-3.5" />
          {books.length}
          {books.length === 1 ? 'book' : 'books'}
        </p>
      </div>
      <div class="flex shrink-0 gap-2">
        <Button
          variant="outline"
          size="sm"
          onclick={() => {
            renameName = author!.name;
            renameOpen = true;
          }}
        >
          <PencilIcon class="size-3.5" /> Rename
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
  {/if}
</div>

<!-- Books section uses BookView which manages its own padding -->
{#if !isLoading && books.length > 0}
  <BookView {books} isLoading={false} emptyMessage="No books by this author." />
{:else if !isLoading && author && books.length === 0}
  <div class="mx-4 flex min-h-32 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20">
    <p class="text-muted-foreground">No books by this author.</p>
  </div>
{/if}

<!-- Rename Dialog -->
<AlertDialog.Root bind:open={renameOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Rename Author</AlertDialog.Title>
      <AlertDialog.Description>
        Enter a new name for "{author?.name}".
      </AlertDialog.Description>
    </AlertDialog.Header>
    <Input bind:value={renameName} placeholder="Author name" />
    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <Button onclick={confirmRename} disabled={renaming || !renameName.trim()}>
        {renaming ? 'Renaming…' : 'Rename'}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<!-- Delete Dialog -->
<AlertDialog.Root bind:open={deleteOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete "{author?.name}"?</AlertDialog.Title>
      <AlertDialog.Description>
        This author will be removed. Their books will not be deleted.
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
      <AlertDialog.Action onclick={confirmDelete} disabled={deleting}>
        {deleting ? 'Deleting…' : 'Delete'}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
