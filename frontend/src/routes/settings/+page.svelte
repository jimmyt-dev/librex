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
  import XIcon from '@lucide/svelte/icons/x';
  import FolderPicker from '$lib/components/folder-picker.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import { buttonVariants } from '$lib/components/ui/button';

  const DEFAULT_PATTERN = '{authors}/{title}{ext}';

  let pattern = $state('');
  let writeMetadata = $state(false);
  let bookdropPath = $state('');
  let folderDialogOpen = $state(false);
  let saving = $state(false);
  let dirty = $derived(
    pattern !== (settingsState.settings?.fileNamingPattern ?? DEFAULT_PATTERN) ||
      writeMetadata !== (settingsState.settings?.writeMetadataToFile ?? false) ||
      bookdropPath !== (settingsState.settings?.bookdropPath ?? '')
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

  function resolveSegment(
    segment: string,
    vars: Record<string, string>
  ): string | null {
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
    });
  });

  async function saveSettings() {
    saving = true;
    const ok = await settingsState.update({
      fileNamingPattern: pattern,
      writeMetadataToFile: writeMetadata,
      bookdropPath: bookdropPath
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
    <h2 class="text-lg font-semibold">Bookdrop</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      The default folder scanned when you open the Bookdrop page.
    </p>
    <div class="mt-4 flex flex-col gap-3">
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium">Bookdrop Folder</label>
        <Button
          type="button"
          variant="outline"
          class="w-full justify-start gap-2 font-normal {!bookdropPath ? 'text-muted-foreground' : ''}"
          onclick={() => (folderDialogOpen = true)}
        >
          {#if bookdropPath}
            <FolderOpenIcon class="size-4 shrink-0" />
            <span class="truncate">{bookdropPath}</span>
            <button
              type="button"
              class="ml-auto text-muted-foreground hover:text-foreground"
              onclick={(e) => { e.stopPropagation(); bookdropPath = ''; }}
            >
              <XIcon class="size-3" />
            </button>
          {:else}
            <FolderIcon class="size-4 shrink-0" />
            Select folder...
          {/if}
        </Button>
      </div>
      <div class="flex justify-end">
        <Button
          onclick={saveSettings}
          disabled={saving || bookdropPath === (settingsState.settings?.bookdropPath ?? '')}
          size="sm"
        >
          {saving ? 'Saving...' : 'Save'}
        </Button>
      </div>
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
        <label for="pattern" class="text-sm font-medium">Default Naming Pattern</label>
        <div class="flex gap-2">
          <Input id="pattern" bind:value={pattern} class="font-mono text-sm" />
          <Button
            variant="outline"
            size="icon"
            onclick={resetPattern}
            title="Reset to default"
          >
            <RotateCcwIcon class="size-4" />
          </Button>
        </div>
      </div>

      <div class="rounded-lg border bg-muted/50 px-3 py-2">
        <p class="text-xs font-medium text-muted-foreground">Preview</p>
        <p class="mt-0.5 font-mono text-sm break-all">{preview}</p>
      </div>

      <label class="flex cursor-pointer items-center gap-2 text-sm">
        <Checkbox bind:checked={writeMetadata} />
        <div>
          <span class="font-medium">Write metadata to book files</span>
          <p class="text-xs text-muted-foreground">
            When enabled, editing a book's metadata will also update the EPUB file itself.
          </p>
        </div>
      </label>

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
            {#each variables as v}
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
