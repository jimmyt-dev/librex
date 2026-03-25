<script lang="ts">
  let {
    value = $bindable('')
  }: {
    value: string;
  } = $props();
</script>

<div class="flex gap-1">
  {#each [1, 2, 3, 4, 5] as star (star)}
    <button
      type="button"
      class="text-2xl leading-none transition-colors {(parseInt(value) || 0) >= star
        ? 'text-yellow-400'
        : 'text-muted-foreground/30'} hover:text-yellow-400"
      onclick={() => (value = value === star.toString() ? '' : star.toString())}>★</button
    >
  {/each}
  {#if value}
    <button
      type="button"
      class="ml-1 self-center text-xs text-muted-foreground hover:text-foreground"
      onclick={() => (value = '')}>Clear</button
    >
  {/if}
</div>
