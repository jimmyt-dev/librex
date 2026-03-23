<script lang="ts">
  import { ScrollArea } from '$lib/components/ui/scroll-area';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
  import SearchIcon from '@lucide/svelte/icons/search';

  let { value = $bindable('') }: { value?: string } = $props();

  let current = $state('/');
  let parent = $state('/');
  let dirs = $state<{ name: string; path: string }[]>([]);
  let loading = $state(false);
  let search = $state('');
  let initialized = $state(false);

  let filtered = $derived(
    search ? dirs.filter((d) => d.name.toLowerCase().includes(search.toLowerCase())) : dirs
  );

  async function browse(path?: string) {
    loading = true;
    try {
      const token = localStorage.getItem('bearer_token') || '';
      const params = path ? `?path=${encodeURIComponent(path)}` : '';
      const res = await fetch(`/api/directories${params}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        current = data.current;
        parent = data.parent;
        dirs = data.dirs;
        value = data.current;
        search = '';
      }
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    if (!initialized) {
      initialized = true;
      browse(value || undefined);
    }
  });
</script>

<div class="rounded-md border">
  <!-- Current path display -->
  <div class="flex items-center gap-2 border-b bg-muted/30 px-3 py-2">
    <FolderOpenIcon class="size-3.5 shrink-0 text-muted-foreground" />
    <span class="truncate text-xs font-medium">{current}</span>
  </div>

  <!-- Controls row -->
  <div class="flex items-center gap-2 border-b px-3 py-2">
    <div class="relative flex-1">
      <SearchIcon class="absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
      <Input bind:value={search} placeholder="Search folders..." class="h-8 pl-7 text-sm" />
    </div>
    <Button
      variant="outline"
      size="icon"
      class="size-8 shrink-0"
      disabled={current === parent}
      onclick={() => browse(parent)}
      title="Go to parent directory"
    >
      <ChevronUpIcon class="size-4" />
    </Button>
  </div>

  <!-- Directory listing -->
  <ScrollArea class="h-52">
    {#if loading}
      <div class="flex items-center justify-center p-8">
        <span class="text-sm text-muted-foreground">Loading...</span>
      </div>
    {:else if filtered.length === 0}
      <div class="flex items-center justify-center p-8">
        <span class="text-sm text-muted-foreground">
          {search ? 'No matching folders' : 'No subdirectories'}
        </span>
      </div>
    {:else}
      <div class="px-1 py-1">
        <div class="mb-1 flex items-center gap-1.5 px-2 pt-1 text-xs text-muted-foreground">
          <FolderIcon class="size-3" />
          <span>{filtered.length} folder{filtered.length !== 1 ? 's' : ''} available</span>
        </div>
        {#each filtered as dir (dir.path)}
          <button
            type="button"
            class="flex w-full items-center gap-2.5 rounded-sm px-2 py-1.5 text-sm hover:cursor-pointer hover:bg-accent"
            onclick={() => browse(dir.path)}
          >
            <FolderIcon class="size-4 shrink-0 text-muted-foreground" />
            <div class="flex flex-col items-start gap-0">
              <span class="text-sm">{dir.name}</span>
              <span class="text-xs text-muted-foreground">{dir.path}</span>
            </div>
            <ChevronRightIcon class="ml-auto size-4 shrink-0 text-muted-foreground" />
          </button>
        {/each}
      </div>
    {/if}
  </ScrollArea>
</div>
