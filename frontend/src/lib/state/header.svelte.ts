class HeaderState {
  title = $state('');
  subtitle = $state<string | null>(null);
}

export const headerState = new HeaderState();
