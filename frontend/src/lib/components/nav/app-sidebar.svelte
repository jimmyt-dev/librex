<script lang="ts" module>
  import AudioWaveformIcon from '@lucide/svelte/icons/audio-waveform';
  import BookOpenIcon from '@lucide/svelte/icons/book-open';
  import BotIcon from '@lucide/svelte/icons/bot';
  import ChartPieIcon from '@lucide/svelte/icons/chart-pie';
  import CommandIcon from '@lucide/svelte/icons/command';
  import FrameIcon from '@lucide/svelte/icons/frame';
  import GalleryVerticalEndIcon from '@lucide/svelte/icons/gallery-vertical-end';
  import HomeIcon from '@lucide/svelte/icons/home';
  import MapIcon from '@lucide/svelte/icons/map';
  import Settings2Icon from '@lucide/svelte/icons/settings-2';
  import SquareTerminalIcon from '@lucide/svelte/icons/square-terminal';
  import BookTextIcon from '@lucide/svelte/icons/book-text';
  import BookCopyIcon from '@lucide/svelte/icons/book-copy';
  import UsersIcon from '@lucide/svelte/icons/users';
  import NotebookPenIcon from '@lucide/svelte/icons/notebook-pen';

  // This is sample data.
  const data = {
    teams: [
      {
        name: 'Acme Inc',
        logo: GalleryVerticalEndIcon,
        plan: 'Enterprise'
      },
      {
        name: 'Acme Corp.',
        logo: AudioWaveformIcon,
        plan: 'Startup'
      },
      {
        name: 'Evil Corp.',
        logo: CommandIcon,
        plan: 'Free'
      }
    ],
    navHome: [
      {
        title: 'Dashboard',
        url: 'dashboard',
        icon: HomeIcon
      },
      {
        title: 'All Books',
        url: 'all-books',
        icon: BookTextIcon
      },
      {
        title: 'Series',
        url: 'series',
        icon: BookCopyIcon
      },
      {
        title: 'Authors',
        url: 'authors',
        icon: UsersIcon
      },
      {
        title: 'Notebook',
        url: 'notebook',
        icon: NotebookPenIcon
      }
    ],
    navMain: [
      {
        title: 'Playground',
        url: '#',
        icon: SquareTerminalIcon,
        isActive: true,
        items: [
          {
            title: 'History',
            url: '#'
          },
          {
            title: 'Starred',
            url: '#'
          },
          {
            title: 'Settings',
            url: '#'
          }
        ]
      },
      {
        title: 'Models',
        url: '#',
        icon: BotIcon,
        items: [
          {
            title: 'Genesis',
            url: '#'
          },
          {
            title: 'Explorer',
            url: '#'
          },
          {
            title: 'Quantum',
            url: '#'
          }
        ]
      },
      {
        title: 'Documentation',
        url: '#',
        icon: BookOpenIcon,
        items: [
          {
            title: 'Introduction',
            url: '#'
          },
          {
            title: 'Get Started',
            url: '#'
          },
          {
            title: 'Tutorials',
            url: '#'
          },
          {
            title: 'Changelog',
            url: '#'
          }
        ]
      },
      {
        title: 'Settings',
        url: '#',
        icon: Settings2Icon,
        items: [
          {
            title: 'General',
            url: '#'
          },
          {
            title: 'Team',
            url: '#'
          },
          {
            title: 'Billing',
            url: '#'
          },
          {
            title: 'Limits',
            url: '#'
          }
        ]
      }
    ],
    projects: [
      {
        name: 'Design Engineering',
        url: '#',
        icon: FrameIcon
      },
      {
        name: 'Sales & Marketing',
        url: '#',
        icon: ChartPieIcon
      },
      {
        name: 'Travel',
        url: '#',
        icon: MapIcon
      }
    ]
  };
</script>

<script lang="ts">
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import type { ComponentProps } from 'svelte';
  import { onMount } from 'svelte';
  import { librariesState } from '$lib/api/libraries.svelte';
  import { shelvesState } from '$lib/api/shelves.svelte';
  import NavHome from './nav-home.svelte';
  import NavLibraries from './nav-libraries.svelte';
  // import NavMain from './nav-main.svelte';
  // import NavProjects from './nav-projects.svelte';
  import NavUser from './nav-user.svelte';
  // import TeamSwitcher from './team-switcher.svelte';
  import NavShelves from './nav-shelves.svelte';
  import LibraryIcon from '@lucide/svelte/icons/library';
  import Separator from '../ui/separator/separator.svelte';

  const sidebar = Sidebar.useSidebar();

  onMount(() => {
    librariesState.fetchAll();
    shelvesState.fetchAll();
  });

  let {
    ref = $bindable(null),
    collapsible = 'icon',
    user,
    ...restProps
  }: ComponentProps<typeof Sidebar.Root> & {
    user: { name: string; email: string; image?: string | null };
  } = $props();
</script>

<Sidebar.Root bind:ref {collapsible} {...restProps}>
  <Sidebar.Header>
    <!-- <TeamSwitcher teams={data.teams} /> -->
    {#if sidebar.state === 'collapsed'}
      <LibraryIcon class="h-8 w-8" />
    {:else}
      <h1 class="scroll-m-20 text-4xl font-extrabold tracking-tight lg:text-5xl">Reliquary</h1>
    {/if}
  </Sidebar.Header>
  <Sidebar.Content>
    <NavHome links={data.navHome} />
    <Separator />
    <NavLibraries links={librariesState.items} />
    <Separator />
    <NavShelves links={shelvesState.items} />
    <!-- <NavMain items={data.navMain} />
    <div class="border-t border-border"></div>
    <NavProjects projects={data.projects} /> -->
  </Sidebar.Content>
  <Sidebar.Footer>
    <NavUser {user} />
  </Sidebar.Footer>
  <Sidebar.Rail />
</Sidebar.Root>
