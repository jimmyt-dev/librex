<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import PlusIcon from '@lucide/svelte/icons/plus';

  let { onAdd }: { onAdd: (name: string, icon?: string) => Promise<void> } = $props();

  let open = $state(false);
  let name = $state('');
  let icon = $state('');
  let booksFolder = $state('');
  let loading = $state(false);

  async function handleSubmit() {
    if (!name.trim()) return;
    loading = true;
    try {
      await onAdd(name.trim(), icon.trim() || undefined);
      name = '';
      icon = '';
      open = false;
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
        <Label for="library-icon">Icon (optional)</Label>
        <Input id="library-icon" bind:value={icon} placeholder="book-open" />
      </div>
      <div class="col-span-2 grid gap-3">
        <Label for="books-folder">Books Folder</Label>
        <Input id="books-folder" bind:value={booksFolder} placeholder="" />
      </div>
      <Dialog.Footer class="col-span-2">
        <Dialog.Close type="button" class={buttonVariants({ variant: 'outline' })}>
          Cancel
        </Dialog.Close>
        <Button type="submit" disabled={loading || !name.trim()}>
          {loading ? 'Creating...' : 'Create Library'}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>
