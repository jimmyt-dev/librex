<script lang="ts">
  import { librariesState } from '$lib/api/libraries.svelte';
  import { booksState } from '$lib/api/books.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import * as Tabs from '$lib/components/ui/tabs';
  import * as Select from '$lib/components/ui/select';
  import FileUpload from '$lib/components/file-upload.svelte';
  import { toast } from 'svelte-sonner';
  import UploadIcon from '@lucide/svelte/icons/upload';
  import BuildingLibraryIcon from '@lucide/svelte/icons/library';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import InfoIcon from '@lucide/svelte/icons/info';
  import { Separator } from './ui/separator';

  let open = $state(false);
  let activeTab = $state('library');
  let selectedLibraryId = $state('');

  let libraries = $derived(librariesState.items.filter((l) => l.folder));
  let selectedLibrary = $derived(libraries.find((l) => l.id === selectedLibraryId) ?? null);

  // Auto-select first library when dialog opens
  $effect(() => {
    if (open && !selectedLibraryId && libraries.length > 0) {
      selectedLibraryId = libraries[0].id;
    }
  });

  async function handleLibraryUpload(stagedBooks: unknown[]) {
    // stagedBooks is actually newly added Book[] from the library upload endpoint
    const books = stagedBooks as import('$lib/api/books.svelte').Book[];
    for (const b of books) {
      booksState.upsert(b);
    }
    // Refresh library counts
    librariesState.fetchAll();

    toast.success(
      `${books.length} book${books.length !== 1 ? 's' : ''} added to ${selectedLibrary?.title ?? 'library'}.`
    );
    open = false;
  }

  function handleBookdropUploaded() {
    toast.success('Files uploaded to bookdrop.');
    open = false;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger
    class="inline-flex size-9 items-center justify-center rounded-md border border-input bg-background text-sm font-medium shadow-xs transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-1 focus-visible:ring-ring"
  >
    <UploadIcon class="size-4" />
    <span class="sr-only">Upload books</span>
  </Dialog.Trigger>

  <Dialog.Content class="flex flex-col gap-0 p-0 sm:h-145 sm:max-w-[42vw]">
    <Dialog.Header class="px-6 pt-5 pb-3">
      <Dialog.Title class="text-xl">Upload Books</Dialog.Title>
      <Dialog.Description class="leading-relaxed text-balance">
        Select a destination and drag your ebook files into the drop area below.
      </Dialog.Description>
    </Dialog.Header>

    <Tabs.Root bind:value={activeTab} class="flex flex-1 flex-col overflow-hidden">
      <div class="px-6">
        <Tabs.List variant="line" class="w-full justify-start gap-4">
          <Tabs.Trigger value="library" class="gap-2 pb-3.5 data-[state=active]:text-primary">
            <BuildingLibraryIcon class="size-4" /> Library
          </Tabs.Trigger>
          <Tabs.Trigger value="bookdrop" class="gap-2 pb-3.5 data-[state=active]:text-primary">
            <InboxIcon class="size-4" /> Bookdrop
          </Tabs.Trigger>
        </Tabs.List>
      </div>

      <Separator />

      <div class="flex-1 overflow-y-auto">
        <!-- Library tab -->
        <Tabs.Content value="library" class="m-0 flex flex-col gap-5 p-6">
          {#if libraries.length === 0}
            <div class="flex flex-col items-center justify-center gap-3 py-10 text-center">
              <div class="flex size-14 items-center justify-center rounded-full bg-muted/50">
                <BuildingLibraryIcon class="size-7 text-muted-foreground/50" />
              </div>
              <div class="space-y-1">
                <p class="text-sm font-medium">No Library Folders Found</p>
                <p class="max-w-70 text-xs leading-normal text-muted-foreground">
                  You must configure a library with a local folder path in Settings before you can
                  upload files directly to it.
                </p>
              </div>
            </div>
          {:else}
            <div class="space-y-4">
              <div class="space-y-2">
                <label
                  for="upload-library-select"
                  class="text-[10px] font-bold tracking-wider text-muted-foreground uppercase"
                >
                  Target Library
                </label>
                <Select.Root type="single" bind:value={selectedLibraryId}>
                  <Select.Trigger
                    id="upload-library-select"
                    class="h-9 w-full rounded-lg px-3 text-sm"
                  >
                    {selectedLibrary?.title ?? 'Select a destination library'}
                  </Select.Trigger>
                  <Select.Content>
                    {#each libraries as lib (lib.id)}
                      <Select.Item value={lib.id} class="rounded-md">{lib.title}</Select.Item>
                    {/each}
                  </Select.Content>
                </Select.Root>
                {#if selectedLibrary?.folder}
                  <div
                    class="flex items-center gap-2 rounded-md bg-muted/30 px-2 py-1 text-[11px] text-muted-foreground"
                  >
                    <FolderIcon class="size-3 shrink-0 opacity-70" />
                    <span class="truncate">{selectedLibrary.folder}</span>
                  </div>
                {/if}
              </div>

              <div class="pt-1">
                <FileUpload
                  onUploaded={handleLibraryUpload}
                  uploadUrl={selectedLibraryId
                    ? `/api/libraries/${selectedLibraryId}/upload`
                    : undefined}
                />
              </div>
            </div>
          {/if}
        </Tabs.Content>

        <!-- Bookdrop tab -->
        <Tabs.Content value="bookdrop" class="m-0 flex flex-col gap-5 p-6">
          <div class="relative overflow-hidden rounded-xl border bg-muted/30 p-4">
            <div class="relative z-10 flex flex-col gap-2">
              <div class="flex items-center gap-2 text-sm font-semibold text-foreground">
                <div
                  class="flex size-6 items-center justify-center rounded-md bg-primary/10 text-primary"
                >
                  <InboxIcon class="size-3.5" />
                </div>
                Bookdrop Staging
              </div>
              <p class="text-[13px] leading-relaxed text-muted-foreground">
                Files uploaded here are placed in your <span class="font-medium text-foreground"
                  >Bookdrop</span
                > directory for later review.
              </p>
              <div class="flex items-center gap-1.5 text-[11px] text-primary/80">
                <InfoIcon class="size-3" />
                <span>Great for bulk uploads or unorganized collections.</span>
              </div>
            </div>
          </div>
          <FileUpload onUploaded={handleBookdropUploaded} />
        </Tabs.Content>
      </div>
    </Tabs.Root>
  </Dialog.Content>
</Dialog.Root>
