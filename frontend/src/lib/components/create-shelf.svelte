<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import IconPicker from '$lib/components/ui/icon-picker.svelte';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import { toast } from 'svelte-sonner';

  let open = $state(false);
  let name = $state('');
  let icon = $state('');
  let loading = $state(false);
  let errorMessage = $state('');

  async function handleSubmit() {
    if (!name.trim()) return;
    loading = true;
    errorMessage = '';
    try {
      await shelvesState.create(name.trim(), icon || undefined);
      toast.success(`Shelf "${name.trim()}" created successfully!`);
      name = '';
      icon = '';
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
      <Dialog.Title>New Shelf</Dialog.Title>
      <Dialog.Description>Create a new shelf to organize your books.</Dialog.Description>
    </Dialog.Header>
    <form
      onsubmit={(e) => {
        e.preventDefault();
        handleSubmit();
      }}
      class="grid grid-cols-2 gap-4"
    >
      <div class="grid gap-3">
        <Label for="shelf-name">Name</Label>
        <Input id="shelf-name" bind:value={name} placeholder="My Shelf" required />
      </div>
      <div class="grid gap-3">
        <Label>Icon (optional)</Label>
        <IconPicker bind:value={icon} />
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
        <Button type="submit" disabled={loading || !name.trim()}>
          {loading ? 'Creating...' : 'Create Shelf'}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>
