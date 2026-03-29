<script lang="ts">
  import { apiFetch } from '$lib/api/client';
  import { toast } from 'svelte-sonner';
  import UploadIcon from '@lucide/svelte/icons/upload';
  import XIcon from '@lucide/svelte/icons/x';
  import { Spinner } from '$lib/components/ui/spinner';

  let {
    onUploaded
  }: {
    onUploaded: (stagedBooks: unknown[]) => void;
  } = $props();

  const VALID_EXTS = new Set(['.epub', '.pdf', '.mobi', '.azw3', '.cbz', '.cbr']);
  const ACCEPT = [...VALID_EXTS].join(',');

  let dragOver = $state(false);
  let uploading = $state(false);
  let selectedFiles = $state<File[]>([]);
  let inputEl = $state<HTMLInputElement | null>(null);

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  function addFiles(files: FileList | null) {
    if (!files) return;
    const valid = Array.from(files).filter((f) => {
      const ext = '.' + f.name.split('.').pop()?.toLowerCase();
      return VALID_EXTS.has(ext);
    });
    const invalid = files.length - valid.length;
    if (invalid > 0) toast.warning(`${invalid} unsupported file${invalid > 1 ? 's' : ''} skipped.`);
    // Deduplicate by name
    const existing = new Set(selectedFiles.map((f) => f.name));
    selectedFiles = [...selectedFiles, ...valid.filter((f) => !existing.has(f.name))];
  }

  function removeFile(name: string) {
    selectedFiles = selectedFiles.filter((f) => f.name !== name);
  }

  async function upload() {
    if (selectedFiles.length === 0) return;
    uploading = true;
    try {
      const formData = new FormData();
      for (const f of selectedFiles) formData.append('files', f);
      const result = await apiFetch('/api/bookdrop/upload', { method: 'POST', body: formData });
      toast.success(`${selectedFiles.length} file${selectedFiles.length > 1 ? 's' : ''} uploaded.`);
      selectedFiles = [];
      onUploaded(result);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Upload failed.');
    } finally {
      uploading = false;
    }
  }
</script>

<!-- Drop zone -->
<div
  role="button"
  tabindex="0"
  class="relative flex min-h-28 cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed transition-colors {dragOver
    ? 'border-primary bg-primary/5'
    : 'border-muted-foreground/25 hover:border-muted-foreground/50'}"
  ondragover={(e) => {
    e.preventDefault();
    dragOver = true;
  }}
  ondragleave={() => (dragOver = false)}
  ondrop={(e) => {
    e.preventDefault();
    dragOver = false;
    addFiles(e.dataTransfer?.files ?? null);
  }}
  onclick={() => inputEl?.click()}
  onkeydown={(e) => e.key === 'Enter' && inputEl?.click()}
>
  <input
    bind:this={inputEl}
    type="file"
    accept={ACCEPT}
    multiple
    class="sr-only"
    onchange={(e) => addFiles((e.currentTarget as HTMLInputElement).files)}
  />
  <UploadIcon class="size-6 text-muted-foreground" />
  <p class="text-sm text-muted-foreground">
    Drop ebook files here or <span class="text-primary">browse</span>
  </p>
  <p class="text-xs text-muted-foreground/60">.epub · .pdf · .mobi · .azw3 · .cbz · .cbr</p>
</div>

<!-- Selected files list -->
{#if selectedFiles.length > 0}
  <div class="mt-3 flex flex-col gap-1.5">
    {#each selectedFiles as file (file.name)}
      <div class="flex items-center justify-between rounded-md border bg-muted/30 px-3 py-1.5">
        <div class="flex min-w-0 flex-1 flex-col">
          <p class="truncate text-sm font-medium">{file.name}</p>
          <p class="text-xs text-muted-foreground">{formatSize(file.size)}</p>
        </div>
        <button
          type="button"
          class="ml-2 shrink-0 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          onclick={() => removeFile(file.name)}
        >
          <XIcon class="size-3.5" />
        </button>
      </div>
    {/each}

    <button
      type="button"
      class="mt-1 flex w-full items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
      disabled={uploading}
      onclick={upload}
    >
      {#if uploading}
        <Spinner class="size-4" /> Uploading…
      {:else}
        <UploadIcon class="size-4" />
        Upload {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''}
      {/if}
    </button>
  </div>
{/if}
