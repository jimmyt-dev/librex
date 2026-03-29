<script lang="ts">
  import { booksState, type Book } from '$lib/api/books.svelte';
  import StarRating from '$lib/components/star-rating.svelte';
  import { Label } from '$lib/components/ui/label';
  import { Input } from '$lib/components/ui/input';
  import { toast } from 'svelte-sonner';

  let {
    book,
    compact = false
  }: {
    book: Book;
    compact?: boolean;
  } = $props();

  function toDateInput(d: string | null | undefined): string {
    if (!d) return '';
    return d.slice(0, 10);
  }

  let localStatus = $state('unread');
  let localProgress = $state(0);
  let localRating = $state('');
  let localDateStarted = $state('');
  let localDateFinished = $state('');

  // Sync from book whenever progress changes
  $effect(() => {
    localStatus = book.progress?.status ?? 'unread';
    localProgress = book.progress?.progress ?? 0;
    localRating = String(book.progress?.personalRating ?? '');
    localDateStarted = toDateInput(book.progress?.dateStarted);
    localDateFinished = toDateInput(book.progress?.dateFinished);
  });

  // Watch rating changes (StarRating uses bind:value)
  $effect(() => {
    const r = localRating;
    const current = String(book.progress?.personalRating ?? '');
    if (r !== current) {
      const v = r === '' ? undefined : parseFloat(r);
      booksState.updateProgress(book.id, { personalRating: v }).catch(() => {
        toast.error('Failed to update rating.');
      });
    }
  });

  async function setStatus(s: string) {
    localStatus = s;
    try {
      await booksState.updateProgress(book.id, { status: s });
    } catch {
      toast.error('Failed to update reading status.');
      localStatus = book.progress?.status ?? 'unread';
    }
  }

  async function setProgress(v: number) {
    try {
      await booksState.updateProgress(book.id, { progress: v });
    } catch {
      toast.error('Failed to update progress.');
    }
  }

  async function setDateStarted(v: string) {
    localDateStarted = v;
    try {
      await booksState.updateProgress(book.id, { dateStarted: v || null });
    } catch {
      toast.error('Failed to update date.');
    }
  }

  async function setDateFinished(v: string) {
    localDateFinished = v;
    try {
      await booksState.updateProgress(book.id, { dateFinished: v || null });
    } catch {
      toast.error('Failed to update date.');
    }
  }
</script>

{#if compact}
  <select
    class="rounded-md border bg-background px-2 py-1 text-xs focus:outline-none"
    value={localStatus}
    onchange={(e) => setStatus((e.currentTarget as HTMLSelectElement).value)}
    onclick={(e) => e.stopPropagation()}
  >
    <option value="unread">Unread</option>
    <option value="reading">Reading</option>
    <option value="finished">Finished</option>
  </select>
{:else}
  <div class="flex flex-col gap-4">
    <div class="grid grid-cols-2 gap-4">
      <div class="flex flex-col gap-1.5">
        <Label class="text-xs text-muted-foreground">Status</Label>
        <select
          class="h-8 rounded-md border bg-background px-2 py-1 text-sm focus:outline-none"
          value={localStatus}
          onchange={(e) => setStatus((e.currentTarget as HTMLSelectElement).value)}
        >
          <option value="unread">Unread</option>
          <option value="reading">Reading</option>
          <option value="finished">Finished</option>
        </select>
      </div>
      <div class="flex flex-col gap-1.5">
        <Label class="text-xs text-muted-foreground">Progress (%)</Label>
        <Input
          type="number"
          min="0"
          max="100"
          class="h-8"
          value={localProgress}
          onchange={(e) => {
            const v = parseInt((e.currentTarget as HTMLInputElement).value);
            if (!isNaN(v)) {
              localProgress = Math.max(0, Math.min(100, v));
              setProgress(localProgress);
            }
          }}
        />
      </div>
    </div>

    <div class="flex flex-col gap-1.5">
      <Label class="text-xs text-muted-foreground">My Rating</Label>
      <StarRating bind:value={localRating} />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="flex flex-col gap-1.5">
        <Label class="text-xs text-muted-foreground">Date Started</Label>
        <Input
          type="date"
          class="h-8"
          value={localDateStarted}
          onchange={(e) => setDateStarted((e.currentTarget as HTMLInputElement).value)}
        />
      </div>
      <div class="flex flex-col gap-1.5">
        <Label class="text-xs text-muted-foreground">Date Finished</Label>
        <Input
          type="date"
          class="h-8"
          value={localDateFinished}
          onchange={(e) => setDateFinished((e.currentTarget as HTMLInputElement).value)}
        />
      </div>
    </div>
  </div>
{/if}
