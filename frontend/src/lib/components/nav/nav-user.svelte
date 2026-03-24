<script lang="ts">
  import * as Avatar from '$lib/components/ui/avatar';
  import { useSidebar } from '$lib/components/ui/sidebar';
  import LogOutIcon from '@lucide/svelte/icons/log-out';
  import * as Dialog from '$lib/components/ui/dialog';
  import { authClient } from '$lib/auth-client';
  import { Button } from '$lib/components/ui/button';

  let { user }: { user: { name: string; email: string; image?: string | null } } = $props();

  const sidebar = useSidebar();

  const initials = $derived(
    user.name
      .split(' ')
      .map((w) => w[0])
      .slice(0, 2)
      .join('')
      .toUpperCase()
  );
</script>

<div class="">
  {#if sidebar.state === 'collapsed'}
    <Dialog.Root>
      <Dialog.Trigger>
        <Avatar.Root class="size-8">
          <Avatar.Image src={user.image ?? undefined} alt={user.name} />
          <Avatar.Fallback>{initials}</Avatar.Fallback>
        </Avatar.Root>
      </Dialog.Trigger>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>Edit Profile</Dialog.Title>
          <Dialog.Description>
            This action cannot be undone. This will permanently delete your account and remove your
            data from our servers.
          </Dialog.Description>
        </Dialog.Header>
      </Dialog.Content>
    </Dialog.Root>
  {:else}
    <div class="flex items-center gap-1">
      <Dialog.Root>
        <Dialog.Trigger
          class="flex min-w-0 flex-2 cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-sidebar-accent"
        >
          <Avatar.Root class="size-8 shrink-0 rounded-lg">
            <Avatar.Image src={user.image ?? undefined} alt={user.name} />
            <Avatar.Fallback>{initials}</Avatar.Fallback>
          </Avatar.Root>
          <div class="grid min-w-0 flex-1 text-start text-sm leading-tight">
            <span class="truncate font-medium">{user.name}</span>
            <span class="truncate text-xs text-sidebar-foreground/70">{user.email}</span>
          </div>
        </Dialog.Trigger>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title>Edit Profile</Dialog.Title>
            <Dialog.Description>
              This action cannot be undone. This will permanently delete your account and remove
              your data from our servers.
            </Dialog.Description>
          </Dialog.Header>
        </Dialog.Content>
      </Dialog.Root>

      <Button
        variant="ghost"
        size="icon"
        class="size-12 hover:bg-sidebar-accent!"
        onclick={async () => {
          await authClient.signOut();
          window.location.reload();
        }}
      >
        <LogOutIcon class="size-4" />
        <span class="sr-only">Log out</span>
      </Button>
    </div>
  {/if}
</div>
