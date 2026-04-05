<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import BookIcon from '@lucide/svelte/icons/book';
  import { toast } from 'svelte-sonner';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import InfoIcon from '@lucide/svelte/icons/info';
  import BookOpenTextIcon from '@lucide/svelte/icons/book-open-text';
  import Button from './ui/button/button.svelte';
  import { cn } from '$lib/utils';
  import { Label } from './ui/label';
  import { fade } from 'svelte/transition';
  import BookActionsDropdown from './book-actions-dropdown.svelte';
  import { STATUS_COLORS, STATUS_LABELS } from '$lib/constants/books';

  let {
    book,
    selected = false,
    selectMode = false,
    checkboxes = true,
    onselect
  }: {
    book: Book;
    selected?: boolean;
    selectMode?: boolean;
    checkboxes?: boolean;
    onselect?: (id: string, selected: boolean, shiftKey: boolean) => void;
  } = $props();

  let lastClickShift = false;
  let deleteOpen = $state(false);
  let deleteFile = $state(false);
  let deleting = $state(false);
  let updatingStatus = $state(false);

  function handleCardClick(e: MouseEvent) {
    if (selectMode) {
      onselect?.(book.id, !selected, e.shiftKey);
    }
  }

  async function confirmDelete() {
    deleting = true;
    try {
      await booksState.delete(book.id, deleteFile);
      toast.success(`"${book.metadata.title}" deleted.`);
      deleteOpen = false;
      librariesState.fetchAll();
      shelvesState.fetchAll();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete book.');
    } finally {
      deleting = false;
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="group relative cursor-default outline-none"
  class:!cursor-pointer={selectMode}
  tabindex="0"
  onclick={(e) => handleCardClick(e)}
  transition:fade
  onkeydown={(e) => e.key === 'Enter' && handleCardClick(new MouseEvent('click'))}
>
  {#if checkboxes}
    <button
      class="pointer-events-none absolute top-1.5 left-1.5 z-10 opacity-0 transition-opacity group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100"
      class:opacity-100={selected}
      class:pointer-events-auto={selected}
      onclick={(e) => e.stopPropagation()}
    >
      <Checkbox
        checked={selected}
        onCheckedChange={(v) => onselect?.(book.id, !!v, lastClickShift)}
        onclick={(e) => (lastClickShift = e.shiftKey)}
        class="border-2 border-white bg-white/70 shadow-md drop-shadow-sm backdrop-blur-sm data-[state=checked]:border-primary data-[state=checked]:bg-primary"
        aria-label="Select {book.metadata.title}"
      />
    </button>
  {/if}

  <!-- Series Number Badge -->
  {#if book.metadata.seriesNumber}
    <div
      class="absolute top-1.5 right-1.5 z-10 rounded-full bg-muted px-2 py-0.5 text-xs font-medium"
    >
      {book.metadata.seriesNumber}
    </div>
  {/if}

  <div
    class="flex flex-col overflow-hidden rounded-md border bg-card text-card-foreground shadow-sm transition-all group-focus-within:ring-2 group-focus-within:ring-primary/50 focus:outline-none"
    class:ring-2={selected}
    class:ring-primary={selected}
    class:!ring-primary={selected}
  >
    <div class="relative aspect-2/3 w-full overflow-hidden bg-muted">
      <div
        class="pointer-events-none absolute top-1/2 left-1/2 z-10 flex -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-y-2 opacity-0 transition-opacity group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100"
      >
        <Button
          class="border-accent"
          href="/books/{book.id}"
          size="icon-lg"
          aria-label="View details for {book.metadata.title}"
          onclick={(e) => e.stopPropagation()}
        >
          <InfoIcon class="size-4" />
        </Button>

        <Button
          class="border-accent"
          size="icon-lg"
          aria-label="Read {book.metadata.title}"
          onclick={(e) => {
            e.stopPropagation();
            toast.info('Read functionality coming soon!');
          }}
        >
          <BookOpenTextIcon class="size-4" />
        </Button>
      </div>
      {#if book.progress?.status && book.progress.status !== 'unread' && STATUS_LABELS[book.progress.status]}
        <div
          class="absolute bottom-1.5 left-1.5 z-10 rounded-full px-1.5 py-0.5 text-[10px] font-medium {STATUS_COLORS[
            book.progress.status
          ] ?? 'bg-muted text-muted-foreground'}"
        >
          {STATUS_LABELS[book.progress.status]}
        </div>
      {/if}
      {#if book.metadata.coverPath}
        <img
          src="/api/books/{book.id}/cover"
          alt={book.metadata.title}
          class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
          loading="lazy"
        />
      {:else}
        <div class="flex h-full w-full items-center justify-center text-muted-foreground">
          <BookIcon class="size-8" />
        </div>
      {/if}
    </div>

    <div class="flex min-w-0 items-stretch">
      <Tooltip.Provider delayDuration={400}>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={cn('min-w-0 flex-1 cursor-default px-2 py-2 text-left', {
              'cursor-pointer!': selectMode
            })}
          >
            <p class="truncate text-xs leading-tight font-medium">{book.metadata.title}</p>
            {#if book.authors.length > 0}
              <p class="truncate text-xs text-muted-foreground">
                {book.authors.map((a) => a.name).join(', ')}
              </p>
            {/if}
          </Tooltip.Trigger>
          <Tooltip.Portal>
            <Tooltip.Content side="bottom">{book.metadata.title}</Tooltip.Content>
          </Tooltip.Portal>
        </Tooltip.Root>
      </Tooltip.Provider>

      <BookActionsDropdown
        {book}
        onDelete={() => (deleteOpen = true)}
        class="rounded-none px-1.5"
      />
    </div>
  </div>
</div>

<AlertDialog.Root
  bind:open={deleteOpen}
  onOpenChange={(o) => {
    if (!o) deleteFile = false;
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete "{book.metadata.title}"?</AlertDialog.Title>
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
