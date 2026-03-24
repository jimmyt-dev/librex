export type HeaderCount = {
  label: string;
  value: number;
};

class HeaderState {
  title = $state('');
  subtitle = $state<string | null>(null);
  counts = $state<HeaderCount[]>([]);
}

export const headerState = new HeaderState();
