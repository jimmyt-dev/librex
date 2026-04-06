<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import { bookEditState } from '$lib/state/book-edit.svelte';
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import EllipsisVerticalIcon from '@lucide/svelte/icons/ellipsis-vertical';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import LibraryBigIcon from '@lucide/svelte/icons/library-big';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import TrashIcon from '@lucide/svelte/icons/trash-2';
  import CheckIcon from '@lucide/svelte/icons/check';
  import { toast } from 'svelte-sonner';
  import { STATUS_OPTIONS } from '$lib/constants/books';
  import { cn } from '$lib/utils';

  let {
    book,
    onDelete,
    align = 'start',
    class: className
  }: {
    book: Book;
    onDelete: () => void;
    align?: 'start' | 'end' | 'center';
    class?: string;
  } = $props();

  let updatingStatus = $state(false);

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

  function download() {
    // Using window.location.assign allows the browser to handle the download directly.
    // This shows the download progress immediately and avoids memory issues with large files.
    // Authentication is handled via session cookies.
    window.location.assign(`/api/books/${book.id}/download`);
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    onclick={(e) => e.stopPropagation()}
    class={cn(
      'flex shrink-0 items-center justify-around rounded-none text-muted-foreground hover:bg-muted hover:text-foreground',
      className
    )}
  >
    <EllipsisVerticalIcon class="size-4" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content {align} class="w-44">
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger disabled={updatingStatus}>Read Status</DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent>
          {#each STATUS_OPTIONS as opt (opt.value)}
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                setStatus(opt.value);
              }}
            >
              <CheckIcon
                class="size-3.5 {book.progress?.status === opt.value ? 'opacity-100' : 'opacity-0'}"
              />
              {opt.label}
            </DropdownMenu.Item>
          {/each}
          {#if book.progress?.status}
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              onclick={(e) => {
                e.stopPropagation();
                setStatus(null);
              }}
            >
              <div class="size-3.5"></div>
              Reset Status
            </DropdownMenu.Item>
          {/if}
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>

      <DropdownMenu.Separator />

      <DropdownMenu.Item
        onclick={(e) => {
          e.stopPropagation();
          bookEditState.openFor(book);
        }}
      >
        <PencilIcon class="mr-2 size-3.5" />
        Edit Metadata
      </DropdownMenu.Item>

      <DropdownMenu.Item
        onclick={(e) => {
          e.stopPropagation();
          shelfAssignState.openFor([book.id]);
        }}
      >
        <LibraryBigIcon class="mr-2 size-3.5" />
        Manage Shelves
      </DropdownMenu.Item>

      <DropdownMenu.Item
        onclick={(e) => {
          e.stopPropagation();
          download();
        }}
      >
        <DownloadIcon class="mr-2 size-3.5" />
        Download File
      </DropdownMenu.Item>

      <DropdownMenu.Separator />

      <DropdownMenu.Item
        class="text-destructive focus:text-destructive"
        onclick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
      >
        <TrashIcon class="mr-2 size-3.5" />
        Delete Book
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>
