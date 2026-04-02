<script lang="ts">
  import { booksState } from '$lib/api/books.svelte';
  import { headerState } from '$lib/state/header.svelte';

  headerState.title = 'Reading Stats';
  headerState.subtitle = null;
  headerState.counts = [];

  let isLoading = $state(true);

  $effect(() => {
    isLoading = true;
    booksState.fetchAll().finally(() => (isLoading = false));
  });

  let all = $derived(booksState.all);

  let byStatus = $derived.by(() => {
    const map: Record<string, number> = {};
    for (const b of all) {
      const s = b.progress?.status ?? 'unread';
      map[s] = (map[s] ?? 0) + 1;
    }
    return map;
  });

  let finished = $derived(byStatus['finished'] ?? 0);
  let reading = $derived(byStatus['reading'] ?? 0);
  let rereading = $derived(byStatus['re-reading'] ?? 0);
  let partial = $derived(byStatus['partially-read'] ?? 0);
  let paused = $derived(byStatus['paused'] ?? 0);
  let wontRead = $derived(byStatus['wont-read'] ?? 0);
  let abandoned = $derived(byStatus['abandoned'] ?? 0);
  let unread = $derived(
    (byStatus['unread'] ?? 0) + (all.length - Object.values(byStatus).reduce((a, b) => a + b, 0))
  );

  let activeReading = $derived(reading + rereading + partial + paused);
  let totalPages = $derived(
    all
      .filter((b) => b.progress?.status === 'finished')
      .reduce((s, b) => s + (b.metadata.pageCount ?? 0), 0)
  );

  const STATUS_SEGMENTS = $derived([
    { label: 'Unread', value: unread, color: 'bg-muted-foreground/20' },
    { label: 'Reading', value: reading, color: 'bg-blue-500' },
    { label: 'Re-reading', value: rereading, color: 'bg-blue-400' },
    { label: 'Partial', value: partial, color: 'bg-blue-300' },
    { label: 'Paused', value: paused, color: 'bg-yellow-500' },
    { label: 'Finished', value: finished, color: 'bg-green-500' },
    { label: "Won't Read", value: wontRead, color: 'bg-muted-foreground/40' },
    { label: 'Abandoned', value: abandoned, color: 'bg-red-500' }
  ]);

  let monthlyFinished = $derived.by(() => {
    const now = new Date();
    const months: { label: string; count: number }[] = [];
    for (let i = 11; i >= 0; i--) {
      const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
      const label = d.toLocaleDateString(undefined, { month: 'short', year: '2-digit' });
      const count = all.filter((b) => {
        if (b.progress?.status !== 'finished') return false;
        const df = b.progress.dateFinished;
        if (!df) return false;
        const fd = new Date(df);
        return fd.getFullYear() === d.getFullYear() && fd.getMonth() === d.getMonth();
      }).length;
      months.push({ label, count });
    }
    return months;
  });

  let maxMonthly = $derived(Math.max(1, ...monthlyFinished.map((m) => m.count)));
</script>

<div class="page-content gap-6">
  {#if isLoading}
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <!-- eslint-disable-next-line @typescript-eslint/no-unused-vars -->
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
        <p class="mt-1 text-2xl font-bold">{all.length}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Finished</p>
        <p class="mt-1 text-2xl font-bold text-green-500">{finished}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Currently Reading</p>
        <p class="mt-1 text-2xl font-bold text-blue-500">{activeReading}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Pages Read</p>
        <p class="mt-1 text-2xl font-bold">{totalPages.toLocaleString()}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Unread</p>
        <p class="mt-1 text-2xl font-bold text-muted-foreground">{unread}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Paused</p>
        <p class="mt-1 text-2xl font-bold text-yellow-500">{paused}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Abandoned</p>
        <p class="mt-1 text-2xl font-bold text-red-500">{abandoned}</p>
      </div>
      <div class="rounded-lg border bg-card p-4 shadow-sm">
        <p class="text-xs text-muted-foreground">Won't Read</p>
        <p class="mt-1 text-2xl font-bold text-muted-foreground">{wontRead}</p>
      </div>
    </div>

    <!-- Status Breakdown -->
    <div class="rounded-lg border bg-card p-4 shadow-sm">
      <h2 class="mb-3 text-sm font-semibold">Status Breakdown</h2>
      <div class="flex h-3 w-full overflow-hidden rounded-full">
        {#if all.length === 0}
          <div class="h-full w-full bg-muted"></div>
        {:else}
          {#each STATUS_SEGMENTS as seg (seg.label)}
            {#if seg.value > 0}
              <div
                class="h-full transition-all {seg.color}"
                style="width: {(seg.value / all.length) * 100}%"
                title="{seg.label}: {seg.value}"
              ></div>
            {/if}
          {/each}
        {/if}
      </div>
      <div class="mt-3 flex flex-wrap gap-x-4 gap-y-1.5 text-xs text-muted-foreground">
        {#each STATUS_SEGMENTS as seg (seg.label)}
          {#if seg.value > 0}
            <span class="flex items-center gap-1.5">
              <span class="inline-block h-2.5 w-2.5 rounded-sm {seg.color}"></span>
              {seg.label} ({seg.value})
            </span>
          {/if}
        {/each}
      </div>
    </div>

    <!-- Books Finished Per Month -->
    <div class="rounded-lg border bg-card p-4 shadow-sm">
      <h2 class="mb-4 text-sm font-semibold">Books Finished per Month (Last 12 Months)</h2>
      <div class="flex h-40 items-end gap-1">
        {#each monthlyFinished as m (m.label)}
          <div class="flex flex-1 flex-col items-center gap-1">
            <span class="text-[10px] text-muted-foreground">{m.count > 0 ? m.count : ''}</span>
            <div
              class="w-full rounded-t bg-primary"
              style="height: {(m.count / maxMonthly) * 100}%; min-height: {m.count > 0
                ? '4px'
                : '0'};"
            ></div>
          </div>
        {/each}
      </div>
      <div class="mt-1 flex gap-1">
        {#each monthlyFinished as m (m.label)}
          <div class="flex-1 text-center text-[9px] leading-tight text-muted-foreground">
            {m.label}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
