<script lang="ts">
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { buttonVariants } from '$lib/components/ui/button';
  import { Separator } from '$lib/components/ui/separator';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import type { PageData } from './$types';

  let { data }: { data: PageData } = $props();
</script>

<Sidebar.Provider>
  <AppSidebar user={data.user} />
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
                <Breadcrumb.Page>Dashboard</Breadcrumb.Page>
              </Breadcrumb.Item>
            </Breadcrumb.List>
          </Breadcrumb.Root>
        </div>
        <div class="flex items-center">
          <a href="/bookdrop" class={buttonVariants({ variant: 'outline', size: 'icon' })}>
            <InboxIcon class="size-4" />
          </a>
        </div>
      </div>
    </header>
    <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
      <div class="grid auto-rows-min gap-4 md:grid-cols-3">
        <div class="aspect-video rounded-xl bg-muted/50"></div>
        <div class="aspect-video rounded-xl bg-muted/50"></div>
        <div class="aspect-video rounded-xl bg-muted/50"></div>
      </div>
      <div class="min-h-screen flex-1 rounded-xl bg-muted/50 md:min-h-min"></div>
    </div>
  </Sidebar.Inset>
</Sidebar.Provider>
