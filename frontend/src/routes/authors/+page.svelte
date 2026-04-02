<script lang="ts">
  import { headerState } from '$lib/state/header.svelte';
  import { authorsState, type Author } from '$lib/api/authors.svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import { buttonVariants } from '$lib/components/ui/button';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import UserIcon from '@lucide/svelte/icons/user';
  import BookOpenIcon from '@lucide/svelte/icons/book-open';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import XIcon from '@lucide/svelte/icons/x';
  import SquareCheckBig from '@lucide/svelte/icons/square-check-big';
  import { toast } from 'svelte-sonner';
  import { Input } from '$lib/components/ui/input';
  import Button from '$lib/components/ui/button/button.svelte';

  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);
  let selectedIds = $state<Set<string>>(new Set());
  let lastSelectedId = $state<string | null>(null);

  // Rename dialog
  let renameOpen = $state(false);
  let renameAuthor = $state<Author | null>(null);
  let renameName = $state('');
  let renaming = $state(false);

  // Delete dialog
  let deleteOpen = $state(false);
  let deleting = $state(false);

  let count = $derived(selectedIds.size);
  let authors = $derived(authorsState.items);

  $effect(() => {
    const ids = new Set(authors.map((a) => a.id));
    const pruned = new SvelteSet([...selectedIds].filter((id) => ids.has(id)));
    if (pruned.size !== selectedIds.size) selectedIds = pruned;
  });

  $effect(() => {
    headerState.title = 'Authors';
    headerState.subtitle = null;
    headerState.counts = isLoading ? [] : [{ label: 'authors', value: authors.length }];
  });

  $effect(() => {
    isLoading = true;
    errorMsg = null;
    authorsState
      .fetchAll()
      .catch((e: unknown) => {
        errorMsg = e instanceof Error ? e.message : 'Failed to load authors.';
      })
      .finally(() => {
        isLoading = false;
      });
  });

  function toggleSelect(id: string, sel: boolean, shiftKey: boolean) {
    if (shiftKey && sel && lastSelectedId) {
      const ids = authors.map((a) => a.id);
      const from = ids.indexOf(lastSelectedId);
      const to = ids.indexOf(id);
      const [lo, hi] = from < to ? [from, to] : [to, from];
      const next = new SvelteSet(selectedIds);
      for (let i = lo; i <= hi; i++) next.add(ids[i]);
      selectedIds = next;
    } else {
      const next = new SvelteSet(selectedIds);
      if (sel) next.add(id);
      else next.delete(id);
      selectedIds = next;
    }
    if (sel) lastSelectedId = id;
  }

  function handleCardClick(e: MouseEvent, author: Author) {
    if (selectedIds.size > 0) {
      e.preventDefault();
      toggleSelect(author.id, !selectedIds.has(author.id), e.shiftKey);
    }
  }

  function openRename() {
    if (count !== 1) return;
    const id = [...selectedIds][0];
    const author = authors.find((a) => a.id === id);
    if (!author) return;
    renameAuthor = author;
    renameName = author.name;
    renameOpen = true;
  }

  async function confirmRename() {
    if (!renameAuthor || !renameName.trim()) return;
    renaming = true;
    try {
      await authorsState.update(renameAuthor.id, renameName.trim());
      toast.success(`Renamed to "${renameName.trim()}".`);
      renameOpen = false;
      selectedIds = new Set();
    } catch {
      toast.error('Failed to rename author.');
    } finally {
      renaming = false;
    }
  }

  async function confirmDelete() {
    deleting = true;
    const ids = [...selectedIds];
    let failed = 0;
    for (const id of ids) {
      try {
        await authorsState.delete(id);
      } catch {
        failed++;
      }
    }
    deleting = false;
    deleteOpen = false;
    selectedIds = new Set();
    if (failed > 0) {
      toast.error(`Failed to delete ${failed} author${failed > 1 ? 's' : ''}.`);
    } else {
      toast.success(`Deleted ${ids.length} author${ids.length > 1 ? 's' : ''}.`);
    }
  }

  let selectedNames = $derived(
    authors
      .filter((a) => selectedIds.has(a.id))
      .map((a) => a.name)
      .slice(0, 3)
  );
</script>

<div class="page-content gap-4">
  {#if errorMsg}
    <div class="rounded-xl bg-destructive/15 p-4 text-destructive">{errorMsg}</div>
  {/if}

  {#if isLoading}
    <div class="flex min-h-64 items-center justify-center">
      <p class="text-muted-foreground">Loading…</p>
    </div>
  {:else if authors.length === 0}
    <div
      class="flex min-h-64 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
    >
      <p class="text-muted-foreground">No authors yet. Import some books first.</p>
    </div>
  {:else}
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      {#each authors as author (author.id)}
        <a
          href="/authors/{author.id}"
          class="group relative flex flex-col items-center gap-3 overflow-hidden rounded-lg border bg-card p-4 text-card-foreground shadow-sm transition-colors hover:bg-accent"
          class:ring-2={selectedIds.has(author.id)}
          class:ring-primary={selectedIds.has(author.id)}
          onclick={(e) => handleCardClick(e, author)}
        >
          <button
            class="absolute top-1.5 left-1.5 z-10 opacity-0 transition-opacity group-hover:opacity-100"
            class:opacity-100={selectedIds.has(author.id)}
            onclick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              toggleSelect(author.id, !selectedIds.has(author.id), e.shiftKey);
            }}
          >
            <Checkbox
              checked={selectedIds.has(author.id)}
              class="border-2 border-white bg-white/70 shadow-md drop-shadow-sm backdrop-blur-sm data-[state=checked]:border-primary data-[state=checked]:bg-primary"
              aria-label="Select {author.name}"
            />
          </button>
          <div
            class="flex h-16 w-16 items-center justify-center rounded-full bg-muted text-muted-foreground"
          >
            <UserIcon class="size-7" />
          </div>
          <Tooltip.Provider delayDuration={400}>
            <Tooltip.Root>
              <Tooltip.Trigger class="w-full max-w-full min-w-0 text-center">
                <p class="truncate text-sm font-medium">{author.name}</p>
                <p class="flex items-center justify-center gap-1 text-xs text-muted-foreground">
                  <BookOpenIcon class="size-3" />
                  {author.bookCount}
                  {author.bookCount === 1 ? 'book' : 'books'}
                </p>
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content side="bottom">{author.name}</Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
          </Tooltip.Provider>
        </a>
      {/each}
    </div>
  {/if}
</div>

<!-- Selection Toolbar -->
{#if count > 0}
  <div
    class="fixed bottom-6 left-1/2 z-50 flex w-[calc(100vw-2rem)] -translate-x-1/2 items-center justify-between rounded-lg border bg-card px-3 py-2 shadow-lg sm:w-xl md:w-2xl lg:w-3xl"
  >
    <div class="flex items-center">
      <span class="text-sm font-medium">{count} selected</span>
    </div>

    <div class="flex items-center gap-x-2">
      <!-- Select All -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={() => (selectedIds = new SvelteSet(authors.map((a) => a.id)))}
          >
            <SquareCheckBig class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Select All</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <!-- Clear Selection -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={() => (selectedIds = new Set())}
          >
            <XIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Clear Selection</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <div class="h-8 w-px bg-border"></div>

      <!-- Rename (single selection only) -->
      {#if count === 1}
        <Tooltip.Provider delayDuration={400}>
          <Tooltip.Root>
            <Tooltip.Trigger
              class={buttonVariants({ variant: 'outline', size: 'icon' })}
              onclick={openRename}
            >
              <PencilIcon class="size-4" />
            </Tooltip.Trigger>
            <Tooltip.Portal>
              <Tooltip.Content>Rename Author</Tooltip.Content>
            </Tooltip.Portal>
          </Tooltip.Root>
        </Tooltip.Provider>
      {/if}

      <!-- Delete -->
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'destructive', size: 'icon' })}
            onclick={() => (deleteOpen = true)}
          >
            <TrashIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content>Delete Selected</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>
    </div>
    <div></div>
  </div>
{/if}

<!-- Rename Dialog -->
<AlertDialog.Root bind:open={renameOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Rename Author</AlertDialog.Title>
      <AlertDialog.Description>
        Enter a new name for "{renameAuthor?.name}".
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
      <AlertDialog.Title>Delete {count} author{count > 1 ? 's' : ''}?</AlertDialog.Title>
      <AlertDialog.Description>
        {#if selectedNames.length < count}
          "{selectedNames.join('", "')}" and {count - selectedNames.length} more will be removed. Their
          books will not be deleted.
        {:else}
          "{selectedNames.join('", "')}" will be removed. Their books will not be deleted.
        {/if}
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
