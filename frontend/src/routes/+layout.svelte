<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';
  import { ModeWatcher } from 'mode-watcher';
  import { Toaster } from '$lib/components/ui/sonner/index.js';
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Separator } from '$lib/components/ui/separator';
  import type { LayoutData } from './$types';
  import { headerState } from '$lib/state/header.svelte';
  import './layout.css';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import { buttonVariants } from '$lib/components/ui/button';

  let { data, children }: { data: LayoutData; children: import('svelte').Snippet } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<ModeWatcher />
<Toaster richColors />

<Sidebar.Provider open={data.sidebarOpen}>
  {#if data.user}
    <AppSidebar user={data.user} />
  {/if}
  <Sidebar.Inset>
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
        </div>
        <div>
          <a href="/bookdrop" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
            <InboxIcon class="size-4" />
          </a>
        </div>
      </div>
    </header>
    {@render children()}
  </Sidebar.Inset>
</Sidebar.Provider>
