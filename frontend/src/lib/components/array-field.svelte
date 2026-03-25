<script lang="ts">
  import TagInput from './tag-input.svelte';
  import Label from './ui/label/label.svelte';

  let {
    label,
    values = $bindable<string[]>([]),
    mode = $bindable<'merge' | 'replace'>('merge'),
    placeholder = '',
    fetchSuggestions
  }: {
    label: string;
    values: string[];
    mode: 'merge' | 'replace';
    placeholder?: string;
    fetchSuggestions?: (q: string) => Promise<string[]>;
  } = $props();
</script>

<div class="flex flex-col gap-1.5">
  <div class="flex items-center justify-between">
    <Label class="text-sm font-medium">{label}</Label>
    <div class="flex gap-1">
      {#each ['merge', 'replace'] as const as m (m)}
        <button
          type="button"
          class="rounded px-2 py-0.5 text-xs capitalize {mode === m
            ? 'bg-primary text-primary-foreground'
            : 'text-muted-foreground hover:text-foreground'}"
          onclick={() => (mode = m)}>{m}</button
        >
      {/each}
    </div>
  </div>
  <TagInput bind:values {placeholder} {fetchSuggestions} />
</div>
