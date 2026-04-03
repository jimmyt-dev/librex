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
  import { bookdropState } from '$lib/api/bookdrop.svelte';
  import BookEditSheet from '$lib/components/book-edit-sheet.svelte';
  import ShelfAssignDialog from '$lib/components/shelf-assign-dialog.svelte';
  import UploadDialog from '$lib/components/upload-dialog.svelte';
  import GlobalSearch from '$lib/components/global-search.svelte';
  import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
  import * as Tooltip from '$lib/components/ui/tooltip';

  let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();

  let isAuthPage = $derived(page.url.pathname === '/login' || page.url.pathname === '/register');
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<ModeWatcher />
<Toaster richColors closeButton position="top-center" />
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
        class="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
      >
        <div class="flex w-full items-center gap-2 px-4">
          <div class="flex shrink-0 items-center">
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
          <div class="ml-3">
            <GlobalSearch />
          </div>
          <div class="flex flex-1"></div>
          <Tooltip.Provider delayDuration={400}>
            <div class="flex shrink-0 items-center gap-1.5">
              <Tooltip.Root>
                <Tooltip.Trigger>
                  <a href="/stats" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
                    <BarChart3Icon class="size-4" />
                  </a>
                </Tooltip.Trigger>
                <Tooltip.Content>Reading Stats</Tooltip.Content>
              </Tooltip.Root>

              <UploadDialog />

              <Tooltip.Root>
                <Tooltip.Trigger>
                  <a href="/settings" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
                    <SettingsIcon class="size-4" />
                  </a>
                </Tooltip.Trigger>
                <Tooltip.Content>Settings</Tooltip.Content>
              </Tooltip.Root>

              <Tooltip.Root>
                <Tooltip.Trigger>
                  <a
                    href="/bookdrop"
                    class="relative {buttonVariants({ variant: 'outline', size: 'icon' })}"
                  >
                    <InboxIcon class="size-4" />
                    {#if bookdropState.stagedCount > 0}
                      <span
                        class="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-0.5 text-[10px] font-bold text-primary-foreground"
                        >{bookdropState.stagedCount > 99 ? '99+' : bookdropState.stagedCount}</span
                      >
                    {/if}
                  </a>
                </Tooltip.Trigger>
                <Tooltip.Content>Bookdrop</Tooltip.Content>
              </Tooltip.Root>
            </div>
          </Tooltip.Provider>
        </div>
      </header>
      {#key page.url.pathname}
        {@render children()}
      {/key}
    </Sidebar.Inset>
  </Sidebar.Provider>
{/if}
