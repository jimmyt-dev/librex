<script lang="ts">
  import favicon from '$lib/assets/favicon.svg';
  import { ModeWatcher } from 'mode-watcher';
  import { Toaster } from '$lib/components/ui/sonner/index.js';
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import type { LayoutData } from './$types';
  import './layout.css';

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
    {@render children()}
  </Sidebar.Inset>
</Sidebar.Provider>
