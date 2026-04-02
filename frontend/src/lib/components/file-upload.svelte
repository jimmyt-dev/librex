<script lang="ts">
  import { getToken } from '$lib/api/client';
  import { toast } from 'svelte-sonner';
  import UploadIcon from '@lucide/svelte/icons/upload';
  import XIcon from '@lucide/svelte/icons/x';
  import CheckIcon from '@lucide/svelte/icons/check';
  import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
  import { Spinner } from '$lib/components/ui/spinner';

  let {
    onUploaded,
    uploadUrl,
    maxFileSizeMB = 500
  }: {
    onUploaded: (result: unknown[]) => void;
    uploadUrl?: string;
    maxFileSizeMB?: number;
  } = $props();

  const VALID_EXTS = new Set(['.epub', '.pdf', '.mobi', '.azw3', '.cbz', '.cbr']);
  const ACCEPT = [...VALID_EXTS].join(',');

  let dragOver = $state(false);
  let uploading = $state(false);
  let selectedFiles = $state<File[]>([]);
  let progress = $state<Record<string, number>>({});
  let errors = $state<Record<string, string>>({});
  let inputEl = $state<HTMLInputElement | null>(null);

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  function addFiles(files: FileList | null) {
    if (!files || uploading) return;
    const maxBytes = maxFileSizeMB * 1024 * 1024;

    let invalidExt = 0;
    let tooLarge: string[] = [];
    const valid: File[] = [];

    for (const f of Array.from(files)) {
      const ext = '.' + f.name.split('.').pop()?.toLowerCase();
      if (!VALID_EXTS.has(ext)) {
        invalidExt++;
        continue;
      }
      if (f.size > maxBytes) {
        tooLarge.push(f.name);
        continue;
      }
      valid.push(f);
    }

    if (invalidExt > 0)
      toast.warning(`${invalidExt} unsupported file${invalidExt > 1 ? 's' : ''} skipped.`);
    if (tooLarge.length > 0)
      toast.error(
        `${tooLarge.length} file${tooLarge.length > 1 ? 's' : ''} exceed the ${maxFileSizeMB} MB limit and were skipped.`
      );

    // Deduplicate by name
    const existing = new Set(selectedFiles.map((f) => f.name));
    const newFiles = valid.filter((f) => !existing.has(f.name));
    selectedFiles = [...selectedFiles, ...newFiles];
  }

  function removeFile(name: string) {
    if (uploading) return;
    selectedFiles = selectedFiles.filter((f) => f.name !== name);
    delete progress[name];
    delete errors[name];
  }

  async function uploadFile(file: File, url: string): Promise<unknown[]> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('files', file);

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          progress[file.name] = Math.round((e.loaded / e.total) * 100);
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            resolve(JSON.parse(xhr.responseText));
          } catch {
            resolve([]);
          }
        } else {
          const msg = xhr.responseText || `Error ${xhr.status}`;
          errors[file.name] = msg;
          reject(new Error(msg));
        }
      });

      xhr.addEventListener('error', () => {
        const msg = 'Network error';
        errors[file.name] = msg;
        reject(new Error(msg));
      });

      xhr.open('POST', url);
      const token = getToken();
      if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      xhr.send(formData);
    });
  }

  async function uploadAll() {
    if (selectedFiles.length === 0) return;
    uploading = true;
    errors = {};
    const url = uploadUrl || '/api/bookdrop/upload';
    const allResults: unknown[] = [];
    let failedCount = 0;

    // Sequential upload to keep progress bars meaningful and avoid overwhelming the server
    for (const file of selectedFiles) {
      if (progress[file.name] === 100 && !errors[file.name]) continue;
      try {
        const result = await uploadFile(file, url);
        allResults.push(...(Array.isArray(result) ? result : [result]));
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (e) {
        failedCount++;
      }
    }

    uploading = false;
    if (failedCount === 0) {
      toast.success(`${selectedFiles.length} file${selectedFiles.length > 1 ? 's' : ''} uploaded.`);
      selectedFiles = [];
      progress = {};
      onUploaded(allResults);
    } else if (failedCount < selectedFiles.length) {
      toast.warning(`${selectedFiles.length - failedCount} uploaded, ${failedCount} failed.`);
    } else {
      toast.error('All uploads failed.');
    }
  }
</script>

<!-- Drop zone -->
<div
  role="button"
  tabindex="0"
  class="relative flex min-h-24 cursor-pointer flex-col items-center justify-center gap-1.5 rounded-xl border-2 border-dashed transition-all {dragOver
    ? 'border-primary bg-primary/5 ring-4 ring-primary/10'
    : 'border-muted-foreground/20 bg-muted/5 hover:border-muted-foreground/40 hover:bg-muted/10'} {uploading
    ? 'cursor-not-allowed opacity-60'
    : ''}"
  ondragover={(e) => {
    e.preventDefault();
    if (!uploading) dragOver = true;
  }}
  ondragleave={() => (dragOver = false)}
  ondrop={(e) => {
    e.preventDefault();
    dragOver = false;
    addFiles(e.dataTransfer?.files ?? null);
  }}
  onclick={() => !uploading && inputEl?.click()}
  onkeydown={(e) => e.key === 'Enter' && !uploading && inputEl?.click()}
>
  <input
    bind:this={inputEl}
    type="file"
    accept={ACCEPT}
    multiple
    class="sr-only"
    onchange={(e) => addFiles((e.currentTarget as HTMLInputElement).files)}
  />
  <div class="flex size-10 items-center justify-center rounded-full bg-background shadow-sm">
    <UploadIcon class="size-5 text-muted-foreground" />
  </div>
  <div class="text-center">
    <p class="text-sm font-medium">
      Drop ebooks here or <span class="text-primary underline underline-offset-2">browse</span>
    </p>
    <p class="mt-0.5 text-xs text-muted-foreground/60">Supports EPUB, PDF, MOBI, AZW3, CBZ, CBR</p>
  </div>
</div>

<!-- Selected files list -->
{#if selectedFiles.length > 0}
  <div class="mt-4 flex flex-col gap-2">
    <div class="max-h-[220px] space-y-2 overflow-y-auto pr-1">
      {#each selectedFiles as file (file.name)}
        <div
          class="group relative flex flex-col gap-1.5 rounded-lg border bg-background p-3 shadow-sm transition-all hover:border-primary/30"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 flex-1 flex-col">
              <p class="truncate text-sm leading-none font-medium">{file.name}</p>
              <p
                class="mt-1 text-[10px] font-medium tracking-tight text-muted-foreground uppercase"
              >
                {formatSize(file.size)}
              </p>
            </div>

            <div class="flex items-center gap-2">
              {#if errors[file.name]}
                <div class="text-destructive" title={errors[file.name]}>
                  <AlertCircleIcon class="size-4" />
                </div>
              {:else if progress[file.name] === 100}
                <div class="text-green-600">
                  <CheckIcon class="size-4" />
                </div>
              {:else if uploading && progress[file.name] > 0}
                <span class="text-[10px] font-bold text-primary">{progress[file.name]}%</span>
              {/if}

              <button
                type="button"
                class="rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-30"
                disabled={uploading}
                onclick={() => removeFile(file.name)}
              >
                <XIcon class="size-3.5" />
              </button>
            </div>
          </div>

          <!-- Progress bar -->
          {#if progress[file.name] !== undefined || errors[file.name]}
            <div class="h-1.5 w-full overflow-hidden rounded-full bg-muted">
              <div
                class="h-full transition-all duration-300 {errors[file.name]
                  ? 'bg-destructive'
                  : 'bg-primary'}"
                style="width: {progress[file.name] ?? 100}%"
              ></div>
            </div>
          {/if}

          {#if errors[file.name]}
            <p class="text-[10px] leading-tight text-destructive">{errors[file.name]}</p>
          {/if}
        </div>
      {/each}
    </div>

    <button
      type="button"
      class="mt-2 flex w-full items-center justify-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground shadow-md transition-all hover:bg-primary/90 hover:shadow-lg active:scale-[0.98] disabled:opacity-60"
      disabled={uploading ||
        (selectedFiles.length > 0 &&
          selectedFiles.every((f) => progress[f.name] === 100 && !errors[f.name]))}
      onclick={uploadAll}
    >
      {#if uploading}
        <Spinner class="size-4" />
        <span>Uploading {Object.keys(progress).length} of {selectedFiles.length}...</span>
      {:else}
        <UploadIcon class="size-4" />
        <span>Upload {selectedFiles.length} {selectedFiles.length > 1 ? 'files' : 'file'}</span>
      {/if}
    </button>
  </div>
{/if}
