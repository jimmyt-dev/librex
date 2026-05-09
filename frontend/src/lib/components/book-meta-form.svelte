<script lang="ts">
  import {
    fetchAuthorSuggestions,
    fetchGenreSuggestions,
    fetchPublisherSuggestions,
    fetchSeriesSuggestions,
    fetchTagSuggestions
  } from '$lib/api/suggestions';
  import StarRating from '$lib/components/star-rating.svelte';
  import TagInput from '$lib/components/tag-input.svelte';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  let {
    title = $bindable(''),
    subtitle = $bindable(''),
    authors = $bindable<string[]>([]),
    description = $bindable(''),
    publisher = $bindable(''),
    publishedDate = $bindable(''),
    isbn13 = $bindable(''),
    isbn10 = $bindable(''),
    language = $bindable(''),
    pageCount = $bindable(''),
    seriesName = $bindable(''),
    seriesNumber = $bindable(''),
    seriesTotal = $bindable(''),
    rating = $bindable(''),
    genres = $bindable<string[]>([]),
    tags = $bindable<string[]>([]),
    coverSrc = null as string | null,
    dirtyFields = {} as Record<string, boolean>,
    showIsbn10 = true,
    oninput = () => {}
  }: {
    title?: string;
    subtitle?: string;
    authors?: string[];
    description?: string;
    publisher?: string;
    publishedDate?: string;
    isbn13?: string;
    isbn10?: string;
    language?: string;
    pageCount?: string;
    seriesName?: string;
    seriesNumber?: string;
    seriesTotal?: string;
    rating?: string;
    genres?: string[];
    tags?: string[];
    coverSrc?: string | null;
    dirtyFields?: Record<string, boolean>;
    showIsbn10?: boolean;
    oninput?: () => void;
  } = $props();

  let publisherSuggestions = $state<string[]>([]);
  let showPublisherDropdown = $state(false);
  let publisherHighlightIndex = $state(-1);

  let seriesSuggestions = $state<string[]>([]);
  let showSeriesDropdown = $state(false);
  let seriesHighlightIndex = $state(-1);

  function dirty(field: string) {
    return dirtyFields[field] === true;
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="flex flex-col gap-4" {oninput}>
  {#if coverSrc}
    <img src={coverSrc} alt="Cover" class="mx-auto h-48 w-auto rounded object-contain shadow" />
  {/if}

  <div class="flex flex-col gap-1.5">
    <Label for="bmf-title">
      Title{#if dirty('title')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <Input id="bmf-title" bind:value={title} />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label for="bmf-subtitle">
      Subtitle{#if dirty('subtitle')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <Input id="bmf-subtitle" bind:value={subtitle} />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label>
      Authors{#if dirty('authors')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <TagInput
      bind:values={authors}
      placeholder="Add author…"
      fetchSuggestions={fetchAuthorSuggestions}
    />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label for="bmf-description">
      Description{#if dirty('description')}<span
          class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <Textarea id="bmf-description" bind:value={description} placeholder="Description" />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label for="bmf-publisher">
      Publisher{#if dirty('publisher')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <div class="relative">
      <Input
        id="bmf-publisher"
        bind:value={publisher}
        oninput={async () => {
          publisherHighlightIndex = -1;
          if (publisher.trim().length < 1) {
            publisherSuggestions = [];
            showPublisherDropdown = false;
            return;
          }
          publisherSuggestions = await fetchPublisherSuggestions(publisher.trim());
          showPublisherDropdown = publisherSuggestions.length > 0;
        }}
        onkeydown={(e) => {
          if (e.key === 'ArrowDown') {
            e.preventDefault();
            if (publisherSuggestions.length > 0) {
              showPublisherDropdown = true;
              publisherHighlightIndex = Math.min(
                publisherHighlightIndex + 1,
                publisherSuggestions.length - 1
              );
            }
          } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            publisherHighlightIndex = Math.max(publisherHighlightIndex - 1, -1);
          } else if (e.key === 'Enter' && publisherHighlightIndex >= 0) {
            e.preventDefault();
            publisher = publisherSuggestions[publisherHighlightIndex];
            showPublisherDropdown = false;
            publisherHighlightIndex = -1;
          } else if (e.key === 'Escape') {
            showPublisherDropdown = false;
            publisherHighlightIndex = -1;
          }
        }}
        onfocus={() => {
          if (publisherSuggestions.length > 0) showPublisherDropdown = true;
        }}
        onblur={() => setTimeout(() => (showPublisherDropdown = false), 150)}
      />
      {#if showPublisherDropdown}
        <div
          class="absolute top-full right-0 left-0 z-50 mt-1 max-h-32 overflow-y-auto rounded-lg border bg-popover shadow-md"
        >
          {#each publisherSuggestions as s, i (i)}
            <button
              type="button"
              class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i ===
              publisherHighlightIndex
                ? 'bg-accent'
                : ''}"
              onmousedown={() => {
                publisher = s;
                showPublisherDropdown = false;
                publisherHighlightIndex = -1;
              }}>{s}</button
            >
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <div class="grid grid-cols-2 gap-4">
    <div class="flex flex-col gap-1.5">
      <Label for="bmf-date">
        Published Date{#if dirty('publishedDate')}<span
            class="inline-block size-1.5 rounded-full bg-primary"
          ></span>{/if}
      </Label>
      <Input id="bmf-date" type="date" bind:value={publishedDate} />
    </div>
    <div class="flex flex-col gap-1.5">
      <Label for="bmf-language">
        Language{#if dirty('language')}<span class="inline-block size-1.5 rounded-full bg-primary"
          ></span>{/if}
      </Label>
      <Input id="bmf-language" bind:value={language} placeholder="en" />
    </div>
  </div>

  {#if showIsbn10}
    <div class="grid grid-cols-2 gap-4">
      <div class="flex flex-col gap-1.5">
        <Label for="bmf-isbn13">
          ISBN-13{#if dirty('isbn13')}<span class="inline-block size-1.5 rounded-full bg-primary"
            ></span>{/if}
        </Label>
        <Input id="bmf-isbn13" bind:value={isbn13} />
      </div>
      <div class="flex flex-col gap-1.5">
        <Label for="bmf-isbn10">
          ISBN-10{#if dirty('isbn10')}<span class="inline-block size-1.5 rounded-full bg-primary"
            ></span>{/if}
        </Label>
        <Input id="bmf-isbn10" bind:value={isbn10} />
      </div>
    </div>
  {:else}
    <div class="flex flex-col gap-1.5">
      <Label for="bmf-isbn13">
        ISBN{#if dirty('isbn13')}<span class="inline-block size-1.5 rounded-full bg-primary"
          ></span>{/if}
      </Label>
      <Input id="bmf-isbn13" bind:value={isbn13} />
    </div>
  {/if}

  <div class="flex flex-col gap-1.5">
    <Label for="bmf-page-count">
      Page Count{#if dirty('pageCount')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <Input id="bmf-page-count" type="number" bind:value={pageCount} />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label>
      Series{#if dirty('series')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <div class="grid grid-cols-[1fr_4rem] gap-2">
      <div class="relative">
        <Input
          bind:value={seriesName}
          placeholder="Series name"
          oninput={async () => {
            seriesHighlightIndex = -1;
            if (seriesName.trim().length < 1) {
              seriesSuggestions = [];
              showSeriesDropdown = false;
              return;
            }
            seriesSuggestions = await fetchSeriesSuggestions(seriesName.trim());
            showSeriesDropdown = seriesSuggestions.length > 0;
          }}
          onkeydown={(e) => {
            if (e.key === 'ArrowDown') {
              e.preventDefault();
              if (seriesSuggestions.length > 0) {
                showSeriesDropdown = true;
                seriesHighlightIndex = Math.min(
                  seriesHighlightIndex + 1,
                  seriesSuggestions.length - 1
                );
              }
            } else if (e.key === 'ArrowUp') {
              e.preventDefault();
              seriesHighlightIndex = Math.max(seriesHighlightIndex - 1, -1);
            } else if (e.key === 'Enter' && seriesHighlightIndex >= 0) {
              e.preventDefault();
              seriesName = seriesSuggestions[seriesHighlightIndex];
              showSeriesDropdown = false;
              seriesHighlightIndex = -1;
            } else if (e.key === 'Escape') {
              showSeriesDropdown = false;
              seriesHighlightIndex = -1;
            }
          }}
          onfocus={() => {
            if (seriesSuggestions.length > 0) showSeriesDropdown = true;
          }}
          onblur={() => setTimeout(() => (showSeriesDropdown = false), 150)}
        />
        {#if showSeriesDropdown}
          <div
            class="absolute top-full right-0 left-0 z-50 mt-1 max-h-32 overflow-y-auto rounded-lg border bg-popover shadow-md"
          >
            {#each seriesSuggestions as s, i (i)}
              <button
                type="button"
                class="w-full px-2.5 py-1.5 text-left text-sm hover:bg-accent {i ===
                seriesHighlightIndex
                  ? 'bg-accent'
                  : ''}"
                onmousedown={() => {
                  seriesName = s;
                  showSeriesDropdown = false;
                  seriesHighlightIndex = -1;
                }}>{s}</button
              >
            {/each}
          </div>
        {/if}
      </div>
      <Input type="number" bind:value={seriesNumber} placeholder="#" />
    </div>
    <div class="mt-1 grid grid-cols-2 gap-2">
      <div class="flex flex-col gap-1">
        <Label class="text-xs font-normal text-muted-foreground">Total Books</Label>
        <Input type="number" bind:value={seriesTotal} placeholder="Total" />
      </div>
    </div>
  </div>

  <div class="flex flex-col gap-1.5">
    <Label>
      Rating{#if dirty('rating')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <StarRating bind:value={rating} />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label>
      Genres{#if dirty('genres')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <TagInput
      bind:values={genres}
      placeholder="Add genre…"
      fetchSuggestions={fetchGenreSuggestions}
    />
  </div>

  <div class="flex flex-col gap-1.5">
    <Label>
      Tags{#if dirty('tags')}<span class="inline-block size-1.5 rounded-full bg-primary"
        ></span>{/if}
    </Label>
    <TagInput bind:values={tags} placeholder="Add tag…" fetchSuggestions={fetchTagSuggestions} />
  </div>
</div>
