<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';

  headerState.title = 'Reading Stats';
  headerState.subtitle = null;
  headerState.counts = [];

  let isLoading = $state(true);

  $effect(() => {
    isLoading = true;
    booksState.fetchAll().finally(() => {
      isLoading = false;
    });
  });

  let finished = $derived(booksState.all.filter((b) => b.progress?.status === 'finished'));
  let reading = $derived(booksState.all.filter((b) => b.progress?.status === 'reading'));
  let unread = $derived(
    booksState.all.filter((b) => !b.progress?.status || b.progress.status === 'unread')
  );
  let dnf = $derived(booksState.all.filter((b) => b.progress?.status === 'dnf'));
  let totalPages = $derived(finished.reduce((sum, b) => sum + (b.metadata.pageCount ?? 0), 0));

  let totalBooks = $derived(booksState.all.length);

  let monthlyFinished = $derived.by(() => {
    const now = new Date();
    const months: { label: string; count: number }[] = [];
    for (let i = 11; i >= 0; i--) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
      const label = d.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
      const count = finished.filter((b) => {
        const df = b.progress?.dateFinished;
        if (!df) return false;
        const fd = new Date(df);
        return fd.getFullYear() === d.getFullYear() && fd.getMonth() === d.getMonth();
      }).length;
      months.push({ label, count });
    }
    return months;
  });

  let maxMonthly = $derived(Math.max(1, ...monthlyFinished.map((m) => m.count)));

  let statusTotal = $derived(totalBooks || 1);
</script>

<div class="flex flex-1 flex-col gap-6 p-4 pt-0">
  {#if isLoading}
    <!-- Skeleton -->
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
      {#each Array(4) as _, i (i)}
        <div class="rounded-lg border bg-card p-4 shadow-sm">
          <div class="mb-2 h-3 w-1/2 animate-pulse rounded bg-muted"></div>
          <div class="h-7 w-1/3 animate-pulse rounded bg-muted"></div>
        </div>
      {/each}
    </div>
  {:else}
    <!-- Stat Cards -->
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Total Books</p>
        <p class="mt-1 text-2xl font-bold">{totalBooks}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Finished</p>
        <p class="mt-1 text-2xl font-bold text-green-500">{finished.length}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Reading</p>
        <p class="mt-1 text-2xl font-bold text-blue-500">{reading.length}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Pages Read</p>
        <p class="mt-1 text-2xl font-bold">{totalPages.toLocaleString()}</p>
      </div>
    </div>

    <!-- Status Breakdown -->
    <div class="rounded-lg border bg-card p-4 shadow-sm">
      <h2 class="mb-3 text-sm font-semibold">Status Breakdown</h2>
      <div class="flex h-3 w-full overflow-hidden rounded-full">
        {#if totalBooks === 0}
          <div class="h-full w-full bg-muted"></div>
        {:else}
          <div
            class="h-full bg-muted transition-all"
            style="width: {(unread.length / statusTotal) * 100}%"
            title="Unread: {unread.length}"
          ></div>
          <div
            class="h-full bg-blue-500 transition-all"
            style="width: {(reading.length / statusTotal) * 100}%"
            title="Reading: {reading.length}"
          ></div>
          <div
            class="h-full bg-green-500 transition-all"
            style="width: {(finished.length / statusTotal) * 100}%"
            title="Finished: {finished.length}"
          ></div>
          <div
            class="h-full bg-red-500 transition-all"
            style="width: {(dnf.length / statusTotal) * 100}%"
            title="DNF: {dnf.length}"
          ></div>
        {/if}
      </div>
      <!-- Legend -->
      <div class="mt-3 flex flex-wrap gap-4 text-xs text-muted-foreground">
        <span class="flex items-center gap-1.5">
          <span class="inline-block h-2.5 w-2.5 rounded-sm bg-muted border"></span>
          Unread ({unread.length})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="inline-block h-2.5 w-2.5 rounded-sm bg-blue-500"></span>
          Reading ({reading.length})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="inline-block h-2.5 w-2.5 rounded-sm bg-green-500"></span>
          Finished ({finished.length})
        </span>
        <span class="flex items-center gap-1.5">
          <span class="inline-block h-2.5 w-2.5 rounded-sm bg-red-500"></span>
          DNF ({dnf.length})
        </span>
      </div>
    </div>

    <!-- Books Finished Per Month -->
    <div class="rounded-lg border bg-card p-4 shadow-sm">
      <h2 class="mb-4 text-sm font-semibold">Books Finished per Month (Last 12 Months)</h2>
      <div class="flex h-40 items-end gap-1">
        {#each monthlyFinished as m (m.label)}
          <div class="flex flex-1 flex-col items-center gap-1">
            <span class="text-[10px] text-muted-foreground">{m.count > 0 ? m.count : ''}</span>
            <div class="w-full rounded-t" style="height: {(m.count / maxMonthly) * 100}%; min-height: {m.count > 0 ? '4px' : '0'}; background-color: hsl(var(--primary));"></div>
          </div>
        {/each}
      </div>
      <div class="mt-1 flex gap-1">
        {#each monthlyFinished as m (m.label)}
          <div class="flex-1 text-center text-[9px] text-muted-foreground leading-tight">{m.label}</div>
        {/each}
      </div>
    </div>
  {/if}
</div>
