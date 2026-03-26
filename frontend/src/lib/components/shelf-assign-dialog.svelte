<script lang="ts">
  import { shelfAssignState } from '$lib/state/shelf-assign.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
  import { toast } from 'svelte-sonner';
  import LucideIcon from '$lib/components/lucide-icon.svelte';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import { Label } from './ui/label';
  import { PlusIcon } from '@lucide/svelte';
  import CreateShelf from './create-shelf.svelte';

  function getToken() {
    return localStorage.getItem('bearer_token') || '';
  }

  let loading = $state(true);
  let saving = $state(false);
  let checked = $state<Record<string, boolean>>({});
  let original = $state<Record<string, boolean>>({});
  let openCreateShelf = $state(false);

  $effect(() => {
    if (!shelfAssignState.open || shelfAssignState.bookIds.length === 0) return;
    loading = true;

    const bookIds = shelfAssignState.bookIds;

    Promise.all(
      bookIds.map((id) =>
        fetch(`/api/books/${id}/shelves`, {
          headers: { Authorization: `Bearer ${getToken()}` }
        }).then((r) => (r.ok ? (r.json() as Promise<string[]>) : []))
      )
    ).then((results) => {
      const counts: Record<string, number> = {};
      for (const shelfIds of results) {
        for (const sid of shelfIds) {
          counts[sid] = (counts[sid] || 0) + 1;
        }
      }

      const state: Record<string, boolean> = {};
      for (const shelf of shelvesState.items) {
        state[shelf.id] = (counts[shelf.id] || 0) === bookIds.length;
      }
      checked = { ...state };
      original = { ...state };
      loading = false;
    });
  });

  let hasChanges = $derived(shelvesState.items.some((s) => checked[s.id] !== original[s.id]));

  async function save() {
    saving = true;
    const bookIds = shelfAssignState.bookIds;

    try {
      for (const shelf of shelvesState.items) {
        if (checked[shelf.id] && !original[shelf.id]) {
          await shelvesState.addBooks(shelf.id, bookIds);
        } else if (!checked[shelf.id] && original[shelf.id]) {
          await shelvesState.removeBooks(shelf.id, bookIds);
        }
      }
      toast.success('Shelves updated.');
      shelfAssignState.close();
    } catch {
      toast.error('Failed to update shelves.');
    } finally {
      saving = false;
    }
  }
</script>

<Dialog.Root
  bind:open={shelfAssignState.open}
  onOpenChange={(o) => {
    if (!o) {
      checked = {};
      original = {};
      loading = true;
      saving = false;
    }
  }}
>
  <Dialog.Content class="sm:max-w-sm">
    <Dialog.Header>
      <Dialog.Title>
        {shelfAssignState.bookIds.length === 1
          ? 'Manage Shelves'
          : `Shelve ${shelfAssignState.bookIds.length} Books`}
      </Dialog.Title>
      <Dialog.Description>
        {shelfAssignState.bookIds.length === 1
          ? 'Select which shelves this book belongs to.'
          : 'Select shelves to add or remove these books.'}
      </Dialog.Description>
    </Dialog.Header>

    {#if loading}
      <div class="flex items-center justify-center py-8">
        <p class="text-sm text-muted-foreground">Loading…</p>
      </div>
    {:else if shelvesState.items.length === 0}
      <div class="flex items-center justify-center py-8">
        <p class="text-sm text-muted-foreground">No shelves yet. Create one first.</p>
      </div>
    {:else}
      <div class="flex flex-col gap-1 py-2">
        {#each shelvesState.items as shelf (shelf.id)}
          <Label
            class="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm hover:bg-muted"
          >
            <Checkbox
              checked={checked[shelf.id] ?? false}
              onCheckedChange={(v) => (checked[shelf.id] = !!v)}
              class="size-4 accent-primary"
            />
            {#if shelf.icon}
              <LucideIcon name={shelf.icon} class="size-4 text-muted-foreground" />
            {/if}
            <span>{shelf.title}</span>
            <span class="ml-auto text-xs text-muted-foreground">{shelf.books}</span>
          </Label>
        {/each}
      </div>
    {/if}

    <Dialog.Footer class="flex justify-between!">
      <div>
        <Button variant="outline" onclick={() => (openCreateShelf = true)}>
          <PlusIcon class="size-4 text-muted-foreground" />
          Create Shelf
        </Button>
      </div>
      <div>
        <Dialog.Close>
          {#snippet child({ props })}
            <Button variant="outline" {...props}>Cancel</Button>
          {/snippet}
        </Dialog.Close>
        <Button onclick={save} disabled={saving || !hasChanges}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<CreateShelf open={openCreateShelf} onClose={() => (openCreateShelf = false)} />
