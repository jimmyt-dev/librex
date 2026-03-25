<script lang="ts">
  import { cn } from '$lib/utils';
  import XIcon from '@lucide/svelte/icons/x';

  let {
    values = $bindable([]),
    placeholder = '',
    fetchSuggestions
  }: {
    values: string[];
    placeholder?: string;
    fetchSuggestions?: (query: string) => Promise<string[]>;
  } = $props();

  let inputValue = $state('');
  let suggestions = $state<string[]>([]);
  let showSuggestions = $state(false);
  let highlightedIndex = $state(-1);
  let inputEl = $state<HTMLInputElement | null>(null);
  let debounceTimer: ReturnType<typeof setTimeout>;

  function addValues(input: string) {
    const parts = input
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s && !values.includes(s));
    if (parts.length > 0) {
      values = [...values, ...parts];
    }
    inputValue = '';
    suggestions = [];
    showSuggestions = false;
    highlightedIndex = -1;
  }

  function removeValue(val: string) {
    values = values.filter((v) => v !== val);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      if (highlightedIndex >= 0 && highlightedIndex < suggestions.length) {
        addValues(suggestions[highlightedIndex]);
      } else if (inputValue.trim()) {
        addValues(inputValue);
      }
    } else if (e.key === 'Backspace' && inputValue === '' && values.length > 0) {
      values = values.slice(0, -1);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (suggestions.length > 0) {
        showSuggestions = true;
        highlightedIndex = Math.min(highlightedIndex + 1, suggestions.length - 1);
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      highlightedIndex = Math.max(highlightedIndex - 1, -1);
    } else if (e.key === 'Escape') {
      showSuggestions = false;
      highlightedIndex = -1;
    }
  }

  function handlePaste(e: ClipboardEvent) {
    const text = e.clipboardData?.getData('text') ?? '';
    if (text.includes(',')) {
      e.preventDefault();
      addValues(text);
    }
  }

  async function handleInput() {
    highlightedIndex = -1;
    if (!fetchSuggestions || inputValue.trim().length < 1) {
      suggestions = [];
      showSuggestions = false;
      return;
    }
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(async () => {
      const results = await fetchSuggestions(inputValue.trim());
      suggestions = results.filter((r) => !values.includes(r));
      showSuggestions = suggestions.length > 0;
    }, 200);
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="relative">
  <div
    class={cn(
      'flex min-h-8 flex-wrap items-center gap-1 rounded-lg border border-input bg-transparent px-2 py-1 text-sm transition-colors',
      'focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/50',
      'dark:bg-input/30'
    )}
    onclick={() => inputEl?.focus()}
  >
    {#each values as val (val)}
      <span
        class="inline-flex items-center gap-0.5 rounded-md bg-muted px-1.5 py-0.5 text-xs font-medium"
      >
        {val}
        <button
          type="button"
          class="ml-0.5 inline-flex items-center rounded-sm hover:text-destructive"
          onclick={() => removeValue(val)}
        >
          <XIcon class="size-3" />
        </button>
      </span>
    {/each}
    <input
      bind:this={inputEl}
      bind:value={inputValue}
      oninput={handleInput}
      onkeydown={handleKeydown}
      onpaste={handlePaste}
      onfocus={() => {
        if (suggestions.length > 0) showSuggestions = true;
      }}
      onblur={() => {
        // Delay to allow click on suggestion
        setTimeout(() => {
          showSuggestions = false;
        }, 150);
      }}
      {placeholder}
      class="min-w-16 flex-1 border-none bg-transparent text-sm outline-none placeholder:text-muted-foreground"
    />
  </div>

  {#if showSuggestions}
    <div
      class="absolute top-full right-0 left-0 z-50 mt-1 max-h-40 overflow-y-auto rounded-lg border bg-popover shadow-md"
    >
      {#each suggestions as suggestion, i (suggestion)}
        <button
          type="button"
          class={cn(
            'w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent',
            i === highlightedIndex && 'bg-accent'
          )}
          onmousedown={() => addValues(suggestion)}
        >
          {suggestion}
        </button>
      {/each}
    </div>
  {/if}
</div>
