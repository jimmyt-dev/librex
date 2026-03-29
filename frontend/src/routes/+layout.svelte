<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';
  import { ModeWatcher } from 'mode-watcher';
  import { Toaster } from '$lib/components/ui/sonner/index.js';
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Separator } from '$lib/components/ui/separator';
  import { page } from '$app/state';
  import type { LayoutData } from './$types';
  import { headerState } from '$lib/state/header.svelte';
  import './layout.css';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import SettingsIcon from '@lucide/svelte/icons/settings';
  import { buttonVariants } from '$lib/components/ui/button';
  import BookEditSheet from '$lib/components/book-edit-sheet.svelte';
  import ShelfAssignDialog from '$lib/components/shelf-assign-dialog.svelte';
  import UploadDialog from '$lib/components/upload-dialog.svelte';

  let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

  let isAuthPage = $derived(page.url.pathname === '/login' || page.url.pathname === '/register');
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<ModeWatcher />
<Toaster richColors />
<BookEditSheet />
<ShelfAssignDialog />

{#if isAuthPage}
  {@render children()}
{:else}
  <Sidebar.Provider open={data.sidebarOpen}>
    {#if data.user}
      <AppSidebar user={data.user} />
    {/if}
    <Sidebar.Inset class="min-w-0">
      <header
        class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
      >
        <div class="flex w-full items-center justify-between gap-2 px-4">
          <div class="flex items-center">
            <Sidebar.Trigger class="-ms-1" />
            <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
            <Breadcrumb.Root>
              <Breadcrumb.List>
                <Breadcrumb.Item>
                  <Breadcrumb.Page>{headerState.title}</Breadcrumb.Page>
                </Breadcrumb.Item>
              </Breadcrumb.List>
            </Breadcrumb.Root>
            {#if headerState.subtitle}
              <span class="ml-2 text-sm text-muted-foreground">{headerState.subtitle}</span>
            {/if}
            {#if headerState.counts.length > 0}
              <div class="ml-2 flex items-center gap-1.5">
                {#each headerState.counts as count (count.label)}
                  <span class="rounded-md bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                    {count.value}
                    {count.label}
                  </span>
                {/each}
              </div>
            {/if}
          </div>
          <div class="flex items-center gap-1.5">
            <UploadDialog />
            <a href="/settings" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
              <SettingsIcon class="size-4" />
            </a>
            <a href="/bookdrop" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
              <InboxIcon class="size-4" />
            </a>
          </div>
        </div>
      </header>
      {#key page.url.pathname}
        {@render children()}
      {/key}
    </Sidebar.Inset>
  </Sidebar.Provider>
{/if}
