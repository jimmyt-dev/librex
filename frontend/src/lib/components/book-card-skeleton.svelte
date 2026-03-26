<script lang="ts">
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { fade } from 'svelte/transition';
</script>

<div
  class="group relative flex flex-col overflow-hidden rounded-lg border bg-card shadow-sm"
  transition:fade
>
  <div class="aspect-2/3 w-full bg-muted/50">
    <Skeleton class="h-full w-full" />
  </div>
  <div class="flex flex-1 flex-col p-2">
    <Skeleton class="mb-2 h-3 w-3/4" />
    <Skeleton class="h-3 w-1/2" />
  </div>
</div>
