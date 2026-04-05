export const STATUS_OPTIONS = [
  { value: 'unread', label: 'Unread' },
  { value: 'reading', label: 'Reading' },
  { value: 're-reading', label: 'Re-Reading' },
  { value: 'partially-read', label: 'Partially Read' },
  { value: 'paused', label: 'Paused' },
  { value: 'finished', label: 'Read' },
  { value: 'wont-read', label: "Won't Read" },
  { value: 'abandoned', label: 'Abandoned' }
] as const;

export const STATUS_COLORS: Record<string, string> = {
  reading: 'bg-blue-500 text-white',
  'partially-read': 'bg-blue-500 text-white',
  're-reading': 'bg-blue-500 text-white',
  paused: 'bg-yellow-500 text-white',
  finished: 'bg-green-500 text-white',
  'wont-read': 'bg-muted text-muted-foreground',
  abandoned: 'bg-red-500 text-white'
};

export const STATUS_LABELS: Record<string, string> = {
  reading: 'Reading',
  'partially-read': 'Partial',
  're-reading': 'Re-reading',
  paused: 'Paused',
  finished: 'Read',
  'wont-read': "Won't Read",
  abandoned: 'Abandoned'
};
