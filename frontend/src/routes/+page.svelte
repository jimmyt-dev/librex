<script lang="ts">
  import AppSidebar from '$lib/components/nav/app-sidebar.svelte';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb';
  import { Separator } from '$lib/components/ui/separator';
  import * as Sidebar from '$lib/components/ui/sidebar';
  import type { PageData } from './$types';
  import { Button } from '$lib/components/ui/button';

  // Match the StagedBook struct from the Go backend
  interface StagedBook {
    originalPath: string;
    fileName: string;
    title: string;
    ext: string;
  }

  // Svelte 5 reactive state
  let stagedBooks = $state<StagedBook[]>([]);
  let isScanning = $state(false);
  let errorMsg = $state<string | null>(null);

  async function handleScan() {
    isScanning = true;
    errorMsg = null;

    try {
      // Point this to your Go API route.
      // You can add a ?path=/your/custom/path query param if needed.
      const token = localStorage.getItem('bearer_token') || '';
      const res = await fetch('/api/bookdrop/scan', {
        headers: { Authorization: `Bearer ${token}` },
        method: 'POST'
      });

      if (!res.ok) {
        throw new Error('Failed to scan bookdrop');
      }

      stagedBooks = await res.json();
    } catch (err: unknown) {
      errorMsg = err instanceof Error ? err.message : 'An error occurred while scanning.';
      console.error(err);
    } finally {
      isScanning = false;
    }
  }

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
              <Breadcrumb.Item class="hidden md:block">
                <Breadcrumb.Link href="##">Build Your Application</Breadcrumb.Link>
              </Breadcrumb.Item>
              <Breadcrumb.Separator class="hidden md:block" />
              <Breadcrumb.Item>
                <Breadcrumb.Page>Data Fetching</Breadcrumb.Page>
              </Breadcrumb.Item>
            </Breadcrumb.List>
          </Breadcrumb.Root>
        </div>
        <div>
          <Button onclick={handleScan} disabled={isScanning}>
            {isScanning ? 'Scanning...' : 'Scan Bookdrop'}
          </Button>
        </div>
      </div>
    </header>
    <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
      <div class="grid auto-rows-min gap-4 md:grid-cols-3">
        <div class="aspect-video rounded-xl bg-muted/50"></div>
        <div class="aspect-video rounded-xl bg-muted/50"></div>
        <div class="aspect-video rounded-xl bg-muted/50"></div>
      </div>
      <div class="min-h-screen flex-1 rounded-xl bg-muted/50 md:min-h-min">
        <div class="flex flex-1 flex-col gap-4 pt-0">
          {#if errorMsg}
            <div class="rounded-xl bg-destructive/15 p-4 text-destructive">
              {errorMsg}
            </div>
          {/if}

          {#if stagedBooks.length > 0}
            <div class="grid auto-rows-min gap-4 md:grid-cols-3">
              {#each stagedBooks as book (book.title)}
                <div class="flex flex-col gap-2 rounded-xl border bg-muted/50 p-4">
                  <span class="truncate font-medium" title={book.fileName}>
                    {book.fileName}
                  </span>
                  <span class="text-sm text-muted-foreground">
                    Type: {book.ext}
                  </span>
                  <div class="mt-auto flex gap-2 pt-4">
                    <Button size="sm" variant="secondary" class="w-full">Edit & Import</Button>
                  </div>
                </div>
              {/each}
            </div>
          {:else if !isScanning}
            <div
              class="flex min-h-[400px] flex-1 items-center justify-center rounded-xl border-2 border-dashed bg-muted/20"
            >
              <p class="text-muted-foreground">Click "Scan Bookdrop" to find new files.</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </Sidebar.Inset>
</Sidebar.Provider>
