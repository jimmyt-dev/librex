<script lang="ts">
  import { onMount } from 'svelte';
  import { headerState } from '$lib/state/header.svelte';
  import { settingsState } from '$lib/api/settings.svelte';
  import { Input } from '$lib/components/ui/input';
  import { Button } from '$lib/components/ui/button';
  import { Separator } from '$lib/components/ui/separator';
  import { Checkbox } from '$lib/components/ui/checkbox';
  import { toast } from 'svelte-sonner';
  import { setMode, userPrefersMode } from 'mode-watcher';
  import SunIcon from '@lucide/svelte/icons/sun';
  import MoonIcon from '@lucide/svelte/icons/moon';
  import MonitorIcon from '@lucide/svelte/icons/monitor';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import HardDriveIcon from '@lucide/svelte/icons/hard-drive';
  import XIcon from '@lucide/svelte/icons/x';
  import FolderPicker from '$lib/components/folder-picker.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import { buttonVariants } from '$lib/components/ui/button';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { goto } from '$app/navigation';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import LibraryIcon from '@lucide/svelte/icons/library';
  import { dev } from '$app/environment';
  import { Label } from '$lib/components/ui/label';

  const DEFAULT_PATTERN = '{authors}/{title}{ext}';

  let pattern = $state('');
  let writeMetadata = $state(false);
  let bookdropPath = $state('');
  let maxUploadSizeMb = $state(500);
  let folderDialogOpen = $state(false);
  let saving = $state(false);

  // Libraries
  let addLibOpen = $state(false);
  let addLibFolderOpen = $state(false);
  let newLibName = $state('');
  let newLibFolder = $state(dev ? '/' : '/books');
  let newLibPattern = $state('');
  let addingLib = $state(false);
  let deletingLibId = $state<string | null>(null);
  let dirty = $derived(
    pattern !== (settingsState.settings?.fileNamingPattern ?? DEFAULT_PATTERN) ||
      writeMetadata !== (settingsState.settings?.writeMetadataToFile ?? false) ||
      bookdropPath !== (settingsState.settings?.bookdropPath ?? '') ||
      maxUploadSizeMb !== (settingsState.settings?.maxUploadSizeMb ?? 500)
  );

  // Example data for live preview
  const exampleData = {
    authors: 'Patrick Rothfuss',
    title: 'The Name of the Wind',
    series: 'The Kingkiller Chronicle',
    seriesIndex: '01',
    year: '2007',
    publisher: 'DAW Books',
    language: 'en',
    ext: '.epub'
  };

  const variables: { name: string; description: string; example: string }[] = [
    { name: '{title}', description: 'Book title', example: 'The Name of the Wind' },
    { name: '{authors}', description: 'Primary author', example: 'Patrick Rothfuss' },
    { name: '{series}', description: 'Series name', example: 'The Kingkiller Chronicle' },
    { name: '{seriesIndex}', description: 'Series number (zero-padded)', example: '01' },
    { name: '{year}', description: 'Publication year', example: '2007' },
    { name: '{publisher}', description: 'Publisher', example: 'DAW Books' },
    { name: '{language}', description: 'Language code', example: 'en' },
    { name: '{ext}', description: 'File extension', example: '.epub' }
  ];

  function resolvePreview(tmpl: string): string {
    const vars: Record<string, string> = exampleData;
    let result = '';
    let i = 0;
    while (i < tmpl.length) {
      if (tmpl[i] === '<') {
        const end = tmpl.indexOf('>', i + 1);
        if (end === -1) {
          result += tmpl[i];
          i++;
          continue;
        }
        const segment = tmpl.slice(i + 1, end);
        const resolved = resolveSegment(segment, vars);
        if (resolved !== null) result += resolved;
        i = end + 1;
      } else if (tmpl[i] === '{') {
        const end = tmpl.indexOf('}', i + 1);
        if (end === -1) {
          result += tmpl[i];
          i++;
          continue;
        }
        const varName = tmpl.slice(i + 1, end);
        result += vars[varName] ?? 'Unknown';
        i = end + 1;
      } else {
        result += tmpl[i];
        i++;
      }
    }
    return result;
  }

  function resolveSegment(segment: string, vars: Record<string, string>): string | null {
    let result = '';
    let hasEmpty = false;
    let i = 0;
    while (i < segment.length) {
      if (segment[i] === '{') {
        const end = segment.indexOf('}', i + 1);
        if (end === -1) {
          result += segment[i];
          i++;
          continue;
        }
        const varName = segment.slice(i + 1, end);
        if (vars[varName]) {
          result += vars[varName];
        } else {
          hasEmpty = true;
        }
        i = end + 1;
      } else {
        result += segment[i];
        i++;
      }
    }
    return hasEmpty ? null : result;
  }

  let preview = $derived(resolvePreview(pattern));

  onMount(() => {
    headerState.title = 'Settings';
    headerState.subtitle = '';
    headerState.counts = [];
    settingsState.fetch().then(() => {
      pattern = settingsState.settings?.fileNamingPattern ?? DEFAULT_PATTERN;
      writeMetadata = settingsState.settings?.writeMetadataToFile ?? false;
      bookdropPath = settingsState.settings?.bookdropPath ?? '';
      maxUploadSizeMb = settingsState.settings?.maxUploadSizeMb ?? 500;
    });
    librariesState.fetchAll();
  });

  async function addLibrary() {
    if (!newLibName.trim() || !newLibFolder) return;
    addingLib = true;
    try {
      const id = await librariesState.create(
        newLibName.trim(),
        newLibFolder,
        undefined,
        newLibPattern.trim() || undefined
      );
      addLibOpen = false;
      newLibName = '';
      newLibFolder = '/books';
      newLibPattern = '';
      toast.success('Library added. Scanning for books...');
      await librariesState.scan(id);
      toast.success('Scan complete.');
      goto(`/library/${id}`);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to add library.');
    }
    addingLib = false;
  }

  async function deleteLibrary(id: string) {
    deletingLibId = id;
    try {
      await librariesState.delete(id);
      toast.success('Library removed.');
    } catch {
      toast.error('Failed to remove library.');
    }
    deletingLibId = null;
  }

  async function saveSettings() {
    saving = true;
    const ok = await settingsState.update({
      fileNamingPattern: pattern,
      writeMetadataToFile: writeMetadata,
      bookdropPath: bookdropPath,
      maxUploadSizeMb: maxUploadSizeMb
    });
    if (ok) {
      toast.success('Settings saved.');
    } else {
      toast.error('Failed to save settings.');
    }
    saving = false;
  }

  function resetPattern() {
    pattern = DEFAULT_PATTERN;
  }

  let currentTheme = $derived(userPrefersMode.current);
</script>

<div class="mx-auto flex max-w-2xl flex-col gap-8 p-6">
  <!-- Bookdrop -->
  <section>
    <h2 class="text-lg font-semibold">Bookdrop & Uploads</h2>
    <p class="mt-1 text-sm text-muted-foreground">Configure your staging area and upload limits.</p>
    <div class="mt-4 flex flex-col gap-4">
      <div class="flex flex-col gap-1.5">
        <Label class="text-sm font-medium">Bookdrop Folder</Label>
        <Button
          type="button"
          variant="outline"
          class="w-full justify-start gap-2 font-normal {!bookdropPath
            ? 'text-muted-foreground'
            : ''}"
          onclick={() => (folderDialogOpen = true)}
        >
          {#if bookdropPath}
            <FolderOpenIcon class="size-4 shrink-0" />
            <span class="truncate">{bookdropPath}</span>
            <button
              type="button"
              class="ml-auto text-muted-foreground hover:text-foreground"
              onclick={(e) => {
                e.stopPropagation();
                bookdropPath = '';
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

      <div class="flex flex-col gap-1.5">
        <Label for="max-upload" class="text-sm font-medium">Maximum Upload Size (MB)</Label>
        <div class="relative">
          <HardDriveIcon
            class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
          />
          <Input
            id="max-upload"
            type="number"
            bind:value={maxUploadSizeMb}
            min="1"
            class="pl-9"
            placeholder="500"
          />
        </div>
        <p class="text-xs text-muted-foreground">
          Maximum total size for a single upload request. Increase this if you have large ebooks or
          PDFs.
        </p>
      </div>

      <div class="flex justify-end">
        <Button onclick={saveSettings} disabled={saving || !dirty} size="sm">
          {saving ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </div>
  </section>

  <Separator />

  <!-- Libraries -->
  <section>
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">Libraries</h2>
        <p class="mt-1 text-sm text-muted-foreground">
          Manage your book library folders. Each library is a directory of books.
        </p>
      </div>
      <Button size="sm" onclick={() => (addLibOpen = true)}>
        <PlusIcon class="mr-1.5 size-4" />
        Add Library
      </Button>
    </div>

    <div class="mt-4 flex flex-col gap-2">
      {#if librariesState.items.length === 0}
        <p
          class="rounded-lg border border-dashed px-4 py-6 text-center text-sm text-muted-foreground"
        >
          No libraries yet. Add one above.
        </p>
      {:else}
        {#each librariesState.items as lib (lib.id)}
          <div class="flex items-center gap-3 rounded-lg border px-3 py-2.5">
            <LibraryIcon class="size-4 shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium">{lib.title}</p>
              {#if lib.folder}
                <p class="truncate text-xs text-muted-foreground">{lib.folder}</p>
              {/if}
            </div>
            <span class="shrink-0 text-xs text-muted-foreground">{lib.books} books</span>
            <Button
              variant="ghost"
              size="icon"
              class="size-7 shrink-0 text-muted-foreground hover:text-destructive"
              onclick={() => deleteLibrary(lib.id)}
              disabled={deletingLibId === lib.id}
            >
              <Trash2Icon class="size-3.5" />
            </Button>
          </div>
        {/each}
      {/if}
    </div>
  </section>

  <Separator />

  <!-- File Organization -->
  <section>
    <h2 class="text-lg font-semibold">File Organization</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      Configure how imported books are organized on disk. Each library can override this default.
    </p>

    <div class="mt-4 flex flex-col gap-3">
      <div class="flex flex-col gap-1.5">
        <Label for="pattern" class="text-sm font-medium">Default Naming Pattern</Label>
        <div class="flex gap-2">
          <Input id="pattern" bind:value={pattern} class="font-mono text-sm" />
          <Button variant="outline" size="icon" onclick={resetPattern} title="Reset to default">
            <RotateCcwIcon class="size-4" />
          </Button>
        </div>
      </div>

      <div class="rounded-lg border bg-muted/50 px-3 py-2">
        <p class="text-xs font-medium text-muted-foreground">Preview</p>
        <p class="mt-0.5 font-mono text-sm break-all">{preview}</p>
      </div>

      <Label class="flex cursor-pointer items-center gap-2 text-sm">
        <Checkbox bind:checked={writeMetadata} />
        <div>
          <span class="font-medium">Write metadata to book files</span>
          <p class="text-xs text-muted-foreground">
            When enabled, editing a book's metadata will also update the EPUB file itself.
          </p>
        </div>
      </Label>

      <div class="flex justify-end">
        <Button onclick={saveSettings} disabled={saving || !dirty} size="sm">
          {saving ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </div>

    <div class="mt-4">
      <p class="mb-2 text-xs font-medium text-muted-foreground">Available Variables</p>
      <div class="rounded-lg border">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b bg-muted/30">
              <th class="px-3 py-1.5 text-left text-xs font-medium">Variable</th>
              <th class="px-3 py-1.5 text-left text-xs font-medium">Description</th>
              <th class="px-3 py-1.5 text-left text-xs font-medium">Example</th>
            </tr>
          </thead>
          <tbody>
            {#each variables as v (v.name)}
              <tr class="border-b last:border-0">
                <td class="px-3 py-1.5 font-mono text-xs">{v.name}</td>
                <td class="px-3 py-1.5 text-xs text-muted-foreground">{v.description}</td>
                <td class="px-3 py-1.5 text-xs">{v.example}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="mt-2 text-xs text-muted-foreground">
        Wrap optional segments in <code class="rounded bg-muted px-1">&lt;...&gt;</code> — they are
        omitted if the variable is empty. Example:
        <code class="rounded bg-muted px-1">&lt;{'{series'}/&gt;</code>
      </p>
    </div>
  </section>

  <Separator />

  <!-- Appearance -->
  <section>
    <h2 class="text-lg font-semibold">Appearance</h2>
    <p class="mt-1 text-sm text-muted-foreground">Choose your preferred color theme.</p>

    <div class="mt-4 flex gap-2">
      <Button
        variant={currentTheme === 'light' ? 'default' : 'outline'}
        size="sm"
        onclick={() => setMode('light')}
      >
        <SunIcon class="mr-1.5 size-4" />
        Light
      </Button>
      <Button
        variant={currentTheme === 'dark' ? 'default' : 'outline'}
        size="sm"
        onclick={() => setMode('dark')}
      >
        <MoonIcon class="mr-1.5 size-4" />
        Dark
      </Button>
      <Button
        variant={currentTheme === 'system' ? 'default' : 'outline'}
        size="sm"
        onclick={() => setMode('system')}
      >
        <MonitorIcon class="mr-1.5 size-4" />
        System
      </Button>
    </div>
  </section>

  <Separator />

  <!-- Account (placeholder) -->
  <section>
    <h2 class="text-lg font-semibold">Account</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      Account settings will be available here in a future update.
    </p>
  </section>
</div>

<Dialog.Root bind:open={folderDialogOpen}>
  <Dialog.Content class="sm:max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Select Bookdrop Folder</Dialog.Title>
      <Dialog.Description>Choose the folder to scan for new books.</Dialog.Description>
    </Dialog.Header>
    <FolderPicker bind:value={bookdropPath} />
    <Dialog.Footer>
      <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
      <Button onclick={() => (folderDialogOpen = false)} disabled={!bookdropPath}>
        <FolderOpenIcon class="size-4" />
        Select Folder
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Add Library dialog -->
<Dialog.Root bind:open={addLibOpen}>
  <Dialog.Content class="sm:max-w-lg">
    <Dialog.Header>
      <Dialog.Title>Add Library</Dialog.Title>
      <Dialog.Description>Give your library a name and choose its root folder.</Dialog.Description>
    </Dialog.Header>
    <div class="flex flex-col gap-4 py-2">
      <div class="flex flex-col gap-1.5">
        <Label class="text-sm font-medium" for="lib-name">Name</Label>
        <Input id="lib-name" bind:value={newLibName} placeholder="My Library" />
      </div>
      <div class="flex flex-col gap-1.5">
        <Label class="text-sm font-medium">Folder</Label>
        <Button
          type="button"
          variant="outline"
          class="w-full justify-start gap-2 font-normal {!newLibFolder
            ? 'text-muted-foreground'
            : ''}"
          onclick={() => (addLibFolderOpen = true)}
        >
          {#if newLibFolder}
            <FolderOpenIcon class="size-4 shrink-0" />
            <span class="truncate">{newLibFolder}</span>
          {:else}
            <FolderIcon class="size-4 shrink-0" />
            Select folder...
          {/if}
        </Button>
      </div>
      <div class="flex flex-col gap-1.5">
        <Label class="text-sm font-medium" for="lib-pattern">
          Naming Pattern <span class="font-normal text-muted-foreground">(optional)</span>
        </Label>
        <Input
          id="lib-pattern"
          bind:value={newLibPattern}
          placeholder="Leave empty to use default"
          class="font-mono text-sm"
        />
      </div>
    </div>
    <Dialog.Footer>
      <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
      <Button onclick={addLibrary} disabled={addingLib || !newLibName.trim() || !newLibFolder}>
        {addingLib ? 'Adding...' : 'Add Library'}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Add Library folder picker dialog -->
<Dialog.Root bind:open={addLibFolderOpen}>
  <Dialog.Content class="sm:max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>Select Library Folder</Dialog.Title>
      <Dialog.Description>Choose the root folder for this library.</Dialog.Description>
    </Dialog.Header>
    {#key addLibFolderOpen}
      <FolderPicker bind:value={newLibFolder} />
    {/key}
    <Dialog.Footer>
      <Dialog.Close class={buttonVariants({ variant: 'outline' })}>Cancel</Dialog.Close>
      <Button onclick={() => (addLibFolderOpen = false)} disabled={!newLibFolder}>
        <FolderOpenIcon class="size-4" />
        Select Folder
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
