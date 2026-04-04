<script lang="ts">
  import { librariesState, type Library } from '$lib/api/libraries.svelte';
  import { Button, buttonVariants } from '$lib/components/ui/button';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import IconPicker from '$lib/components/icon-picker.svelte';
  import FolderPicker from '$lib/components/folder-picker.svelte';
  import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import { toast } from 'svelte-sonner';

  let {
    open = $bindable(false),
    library
  }: {
    open?: boolean;
    library: Library;
  } = $props();

  let name = $state('');
  let icon = $state('');
  let folder = $state('');
  let fileNamingPattern = $state('');
  let folderDialogOpen = $state(false);
  let loading = $state(false);

  $effect(() => {
    if (open) {
      name = library.title;
      icon = library.icon ?? '';
      folder = library.folder ?? '';
      fileNamingPattern = library.fileNamingPattern ?? '';
    }
  });

  async function handleSubmit() {
    if (!name.trim()) return;
    loading = true;
    try {
      await librariesState.update(library.id, {
        name: name.trim(),
        icon: icon || undefined,
        folder: folder || undefined,
        fileNamingPattern: fileNamingPattern.trim() || undefined
      });
      toast.success('Library updated.');
      open = false;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to update library.');
    } finally {
      loading = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="sm:max-w-xl">
    <Dialog.Header>
      <Dialog.Title>Edit Library</Dialog.Title>
      <Dialog.Description>Update the name, icon, or folder for this library.</Dialog.Description>
    </Dialog.Header>
    <form
      onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}
      class="flex flex-col gap-4 py-2"
    >
      <div class="grid grid-cols-[1fr_auto] gap-4">
        <div class="flex flex-col gap-1.5">
          <Label for="edit-lib-name">Name</Label>
          <Input id="edit-lib-name" bind:value={name} placeholder="My Library" required />
        </div>
        <div class="flex flex-col gap-1.5">
          <Label>Icon</Label>
          <IconPicker bind:value={icon} />
        </div>
      </div>

      <div class="flex flex-col gap-1.5">
        <Label>Folder</Label>
        <Button
          type="button"
          variant="outline"
          class="w-full justify-start gap-2 font-normal {!folder ? 'text-muted-foreground' : ''}"
          onclick={() => (folderDialogOpen = true)}
        >
          {#if folder}
            <FolderOpenIcon class="size-4 shrink-0" />
            <span class="truncate">{folder}</span>
          {:else}
            <FolderIcon class="size-4 shrink-0" />
            Select folder...
          {/if}
        </Button>
      </div>

      <div class="flex flex-col gap-1.5">
        <Label for="edit-lib-pattern">
          File Naming Pattern <span class="font-normal text-muted-foreground">(optional)</span>
        </Label>
        <Input
          id="edit-lib-pattern"
          bind:value={fileNamingPattern}
          placeholder="Leave empty to use default"
          class="font-mono text-sm"
        />
      </div>

      <Dialog.Footer>
        <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
        <Button type="submit" disabled={loading || !name.trim()}>
          {loading ? 'Saving...' : 'Save'}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={folderDialogOpen}>
  <Dialog.Content class="sm:max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Select Library Folder</Dialog.Title>
      <Dialog.Description>Choose the root folder for this library.</Dialog.Description>
    </Dialog.Header>
    {#key folderDialogOpen}
      <FolderPicker bind:value={folder} />
    {/key}
    <Dialog.Footer>
      <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
      <Button onclick={() => (folderDialogOpen = false)} disabled={!folder}>
        <FolderOpenIcon class="size-4" />
        Select Folder
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
