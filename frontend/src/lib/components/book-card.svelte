<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import BookIcon from '@lucide/svelte/icons/book';
  import EllipsisVerticalIcon from '@lucide/svelte/icons/ellipsis-vertical';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import CheckIcon from '@lucide/svelte/icons/check';
  import { toast } from 'svelte-sonner';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import InfoIcon from '@lucide/svelte/icons/info';
  import BookOpenTextIcon from '@lucide/svelte/icons/book-open-text'; // Import Read icon
  import Button from './ui/button/button.svelte';
  import { cn } from '$lib/utils';
  import { Label } from './ui/label';
  import { fade } from 'svelte/transition';

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
  let dropdownOpen = $state(false);
  let updatingStatus = $state(false);

  const STATUS_OPTIONS = [
    { value: 'unread', label: 'Unread' },
    { value: 'reading', label: 'Reading' },
    { value: 're-reading', label: 'Re-Reading' },
    { value: 'partially-read', label: 'Partially Read' },
    { value: 'paused', label: 'Paused' },
    { value: 'finished', label: 'Read' },
    { value: 'wont-read', label: "Won't Read" },
    { value: 'abandoned', label: 'Abandoned' }
  ] as const;

  async function setStatus(status: string | null) {
    if (updatingStatus) return;
    updatingStatus = true;
    try {
      if (status === null) {
        await booksState.deleteProgress(book.id);
      } else {
        await booksState.updateProgress(book.id, { status });
      }
    } catch {
      toast.error('Failed to update reading status.');
    } finally {
      updatingStatus = false;
    }
  }

  const STATUS_COLORS: Record<string, string> = {
    reading: 'bg-blue-500 text-white',
    'partially-read': 'bg-blue-500 text-white',
    're-reading': 'bg-blue-500 text-white',
    paused: 'bg-yellow-500 text-white',
    finished: 'bg-green-500 text-white',
    'wont-read': 'bg-muted text-muted-foreground',
    abandoned: 'bg-red-500 text-white'
  };

  const STATUS_LABELS: Record<string, string> = {
    reading: 'Reading',
    'partially-read': 'Partial',
    're-reading': 'Re-reading',
    paused: 'Paused',
    finished: 'Read',
    'wont-read': "Won't Read",
    abandoned: 'Abandoned'
  };

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
      class="absolute top-1.5 left-1.5 z-10 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 pointer-events-none group-hover:pointer-events-auto group-focus-within:pointer-events-auto"
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
    class="flex flex-col overflow-hidden rounded-md border bg-card text-card-foreground shadow-sm transition-all focus:outline-none group-focus-within:ring-2 group-focus-within:ring-primary/50"
    class:ring-2={selected}
    class:ring-primary={selected}
    class:!ring-primary={selected}
  >
    <div class="relative aspect-2/3 w-full overflow-hidden bg-muted">
      <div
        class="absolute top-1/2 left-1/2 z-10 flex -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-y-2 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 pointer-events-none group-hover:pointer-events-auto group-focus-within:pointer-events-auto"
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

      <DropdownMenu.Root bind:open={dropdownOpen}>
        <DropdownMenu.Trigger
          onclick={(e) => e.stopPropagation()}
          class="flex shrink-0 items-center justify-around rounded-none px-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <EllipsisVerticalIcon class="size-4" />
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content align="start" class="w-44">
            <DropdownMenu.Sub>
              <DropdownMenu.SubTrigger disabled={updatingStatus}
                >Read Status</DropdownMenu.SubTrigger
              >
              <DropdownMenu.SubContent>
                {#each STATUS_OPTIONS as opt (opt.value)}
                  <DropdownMenu.Item
                    onclick={(e) => {
                      e.stopPropagation();
                      setStatus(opt.value);
                    }}
                  >
                    <CheckIcon
                      class="size-3.5 {book.progress?.status === opt.value
                        ? 'opacity-100'
                        : 'opacity-0'}"
                    />
                    {opt.label}
                  </DropdownMenu.Item>
                {/each}
                <DropdownMenu.Separator />
                <DropdownMenu.Item
                  onclick={(e) => {
                    e.stopPropagation();
                    setStatus(null);
                  }}
                >
                  Unset
                </DropdownMenu.Item>
              </DropdownMenu.SubContent>
            </DropdownMenu.Sub>
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                bookEditState.openFor(book);
              }}
            >
              <PencilIcon class="size-3.5" />
              Edit
            </DropdownMenu.Item>
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                shelfAssignState.openFor([book.id]);
              }}
            >
              <LibraryBigIcon class="size-3.5" />
              Shelves
            </DropdownMenu.Item>
            <DropdownMenu.Item
              onclick={async (e) => {
                e.stopPropagation();
                try {
                  const token = localStorage.getItem('bearer_token') || '';
                  const res = await fetch(`/api/books/${book.id}/download`, {
                    headers: { Authorization: `Bearer ${token}` }
                  });
                  if (!res.ok) throw new Error('Download failed');
                  const blob = await res.blob();
                  const disposition = res.headers.get('Content-Disposition') || '';
                  const match = disposition.match(/filename="(.+?)"/);
                  const filename = match ? match[1] : book.metadata.title;
                  const a = document.createElement('a');
                  a.href = URL.createObjectURL(blob);
                  a.download = filename;
                  a.click();
                  URL.revokeObjectURL(a.href);
                } catch {
                  toast.error('Failed to download book.');
                }
              }}
            >
              <DownloadIcon class="size-3.5" />
              Download
            </DropdownMenu.Item>
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              class="text-destructive focus:text-destructive"
              onclick={(e) => {
                e.stopPropagation();
                deleteOpen = true;
              }}
            >
              <TrashIcon class="size-3.5" />
              Delete
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
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
