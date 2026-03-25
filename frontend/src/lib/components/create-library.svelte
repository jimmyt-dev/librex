<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import IconPicker from '$lib/components/icon-picker.svelte';
  import FolderPicker from '$lib/components/folder-picker.svelte';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import XIcon from '@lucide/svelte/icons/x';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { toast } from 'svelte-sonner';

  let open = $state(false);
  let folderDialogOpen = $state(false);
  let name = $state('');
  let icon = $state('');
  let folder = $state('');
  let fileNamingPattern = $state('');
  let loading = $state(false);
  let errorMessage = $state('');

  async function handleSubmit() {
    if (!name.trim() || !folder.trim()) return;
    loading = true;
    errorMessage = '';
    try {
      await librariesState.create(
        name.trim(),
        folder.trim(),
        icon || undefined,
        fileNamingPattern.trim() || undefined
      );
      toast.success(`Library "${name.trim()}" created successfully!`);
      name = '';
      icon = '';
      folder = '';
      fileNamingPattern = '';
      open = false;
    } catch (e) {
      errorMessage = e instanceof Error ? e.message : String(e);
      toast.error(errorMessage);
    } finally {
      loading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger
    type="button"
    class="mx-auto flex items-center justify-center hover:cursor-pointer hover:text-foreground/80"
  >
    <PlusIcon class="size-4" />
  </Dialog.Trigger>
  <Dialog.Content class="sm:max-w-xl">
    <Dialog.Header>
      <Dialog.Title>New Library</Dialog.Title>
      <Dialog.Description>Create a new library to organize your books.</Dialog.Description>
    </Dialog.Header>
    <form
      onsubmit={(e) => {
        e.preventDefault();
        handleSubmit();
      }}
      class="grid grid-cols-2 gap-4"
    >
      <div class="grid gap-3">
        <Label for="library-name">Name</Label>
        <Input id="library-name" bind:value={name} placeholder="My Library" required />
      </div>
      <div class="grid gap-3">
        <Label>Icon (optional)</Label>
        <IconPicker bind:value={icon} />
      </div>
      <div class="col-span-2 grid gap-3">
        <Label>Books Folder</Label>
        <Button
          type="button"
          variant="outline"
          class="w-full justify-start gap-2 font-normal {!folder ? 'text-muted-foreground' : ''}"
          onclick={() => (folderDialogOpen = true)}
        >
          {#if folder}
            <FolderOpenIcon class="size-4 shrink-0" />
            <span class="truncate">{folder}</span>
            <button
              type="button"
              class="ml-auto text-muted-foreground hover:text-foreground"
              onclick={(e) => {
                e.stopPropagation();
                folder = '';
              }}
            >
              <XIcon class="size-3" />
            </button>
          {:else}
            <FolderIcon class="size-4 shrink-0" />
            Select folder...
          {/if}
        </Button>
      </div>

      <div class="col-span-2 grid gap-3">
        <Label for="library-pattern">File Naming Pattern (optional)</Label>
        <Input
          id="library-pattern"
          bind:value={fileNamingPattern}
          placeholder="Leave empty to use default"
          class="font-mono text-sm"
        />
        <p class="text-xs text-muted-foreground">
          Override the default naming pattern for this library. Example: <code class="rounded bg-muted px-1">{'{authors}/{title}{ext}'}</code>
        </p>
      </div>

      {#if errorMessage}
        <div
          class="col-span-2 rounded-md bg-destructive/10 p-3 text-sm font-medium text-destructive"
        >
          {errorMessage}
        </div>
      {/if}

      <Dialog.Footer class="col-span-2">
        <Dialog.Close type="button" class={buttonVariants({ variant: 'outline' })}>
          Cancel
        </Dialog.Close>
        <Button type="submit" disabled={loading || !name.trim() || !folder.trim()}>
          {loading ? 'Creating...' : 'Create Library'}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={folderDialogOpen}>
  <Dialog.Content class="sm:max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Select Directory</Dialog.Title>
      <Dialog.Description>
        Choose a directory from your file system to store books in.
      </Dialog.Description>
    </Dialog.Header>
    <FolderPicker bind:value={folder} />
    <Dialog.Footer>
      <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
      <Button onclick={() => (folderDialogOpen = false)} disabled={!folder}>
        <FolderOpenIcon class="size-4" />
        Select Directory
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
