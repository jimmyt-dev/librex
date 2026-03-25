<script lang="ts">
  import { viewSettings, SORT_FIELDS, type SortField } from '$lib/state/view-settings.svelte';
  import * as Popover from '$lib/components/ui/popover';
  import LayoutGridIcon from '@lucide/svelte/icons/layout-grid';
  import TableIcon from '@lucide/svelte/icons/table';
  import ArrowUpDownIcon from '@lucide/svelte/icons/arrow-up-down';
  import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
  import ArrowDownIcon from '@lucide/svelte/icons/arrow-down';
  import XIcon from '@lucide/svelte/icons/x';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import GripVerticalIcon from '@lucide/svelte/icons/grip-vertical';
  import { Button } from './ui/button';

  let dragIndex = $state<number | null>(null);
  let dropIndex = $state<number | null>(null);

  function onDragStart(e: DragEvent, i: number) {
    dragIndex = i;
    e.dataTransfer!.effectAllowed = 'move';
  }

  function onDragOver(e: DragEvent, i: number) {
    e.preventDefault();
    e.dataTransfer!.dropEffect = 'move';
    dropIndex = i;
  }

  function onDrop(e: DragEvent, i: number) {
    e.preventDefault();
    if (dragIndex !== null && dragIndex !== i) {
      viewSettings.reorderLevels(dragIndex, i);
    }
    dragIndex = null;
    dropIndex = null;
  }

  function onDragEnd() {
    dragIndex = null;
    dropIndex = null;
  }
</script>

<div class="flex w-full items-center gap-2 rounded-md border bg-muted/20 p-2">
  <!-- View toggle -->
  <div class="flex rounded-md border">
    <button
      type="button"
      class="flex items-center gap-1.5 rounded-l-md px-2.5 py-1.5 text-sm transition-colors {viewSettings.mode ===
      'grid'
        ? 'bg-primary text-primary-foreground'
        : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
      onclick={() => viewSettings.setMode('grid')}
      title="Grid view"
    >
      <LayoutGridIcon class="size-4" />
    </button>
    <button
      type="button"
      class="flex items-center gap-1.5 rounded-r-md px-2.5 py-1.5 text-sm transition-colors {viewSettings.mode ===
      'table'
        ? 'bg-primary text-primary-foreground'
        : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
      onclick={() => viewSettings.setMode('table')}
      title="Table view"
    >
      <TableIcon class="size-4" />
    </button>
  </div>

  <!-- Sort -->
  <Popover.Root>
    <Popover.Trigger class="flex items-center gap-1.5">
      {#snippet child({ props })}
        <Button {...props} variant="outline">
          <ArrowUpDownIcon class="size-4" />
          Sort
          {#if viewSettings.sortLevels.length > 0}
            <span
              class="rounded-full border border-background bg-primary px-1.5 py-0 text-xs leading-5 text-primary-foreground"
            >
              {viewSettings.sortLevels.length}
            </span>
          {/if}
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content class="w-72 p-3" align="start">
      <div class="mb-2 flex items-center justify-between">
        <p class="text-sm font-medium">Sort Order</p>
        <button
          type="button"
          class="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
          onclick={() => viewSettings.resetSort()}
        >
          <RotateCcwIcon class="size-3" /> Reset
        </button>
      </div>

      <div class="flex flex-col gap-1">
        {#each viewSettings.sortLevels as level, i (i)}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            aria-label="Sort level {i + 1}: {level.field} {level.dir === 'asc'
              ? 'Ascending'
              : 'Descending'}"
            class="flex items-center gap-1.5 rounded-md px-1 py-0.5 transition-colors {dropIndex ===
              i && dragIndex !== i
              ? 'bg-primary/10 ring-1 ring-primary/30'
              : ''} {dragIndex === i ? 'opacity-40' : ''}"
            draggable="true"
            ondragstart={(e) => onDragStart(e, i)}
            ondragover={(e) => onDragOver(e, i)}
            ondrop={(e) => onDrop(e, i)}
            ondragend={onDragEnd}
            ondragleave={() => {
              if (dropIndex === i) dropIndex = null;
            }}
          >
            <GripVerticalIcon
              class="size-3.5 shrink-0 cursor-grab text-muted-foreground/50 active:cursor-grabbing"
            />

            <span class="w-4 shrink-0 text-right text-xs text-muted-foreground">{i + 1}.</span>

            <select
              class="flex-1 rounded-md border bg-background px-2 py-1 text-xs focus:outline-none"
              value={level.field}
              onchange={(e) =>
                viewSettings.updateLevel(i, {
                  field: (e.currentTarget as HTMLSelectElement).value as SortField
                })}
            >
              {#each SORT_FIELDS as f (f.value)}
                <option value={f.value}>{f.label}</option>
              {/each}
            </select>

            <button
              type="button"
              class="shrink-0 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
              title={level.dir === 'asc' ? 'Ascending' : 'Descending'}
              onclick={() =>
                viewSettings.updateLevel(i, { dir: level.dir === 'asc' ? 'desc' : 'asc' })}
            >
              {#if level.dir === 'asc'}
                <ArrowUpIcon class="size-3.5" />
              {:else}
                <ArrowDownIcon class="size-3.5" />
              {/if}
            </button>

            <button
              type="button"
              class="shrink-0 rounded p-1 text-muted-foreground hover:bg-muted hover:text-destructive"
              onclick={() => viewSettings.removeLevel(i)}
            >
              <XIcon class="size-3.5" />
            </button>
          </div>
        {/each}
      </div>

      {#if viewSettings.sortLevels.length < SORT_FIELDS.length}
        <button
          type="button"
          class="mt-2 w-full rounded-md border border-dashed py-1 text-xs text-muted-foreground transition-colors hover:border-primary hover:text-foreground"
          onclick={() => viewSettings.addLevel()}
        >
          + Add level
        </button>
      {/if}
    </Popover.Content>
  </Popover.Root>
</div>
