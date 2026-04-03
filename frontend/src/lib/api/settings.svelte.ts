import { apiFetch } from './client';

export type UserSettings = {
  id: string;
  userId: string;
  fileNamingPattern: string;
  writeMetadataToFile: boolean;
  maxUploadSizeMb: number;
};

class SettingsState {
  settings = $state<UserSettings | null>(null);
  loading = $state(false);

  async fetch(): Promise<void> {
    this.loading = true;
    try {
      this.settings = await apiFetch('/api/settings');
    } catch (e) {
      console.error('Failed to fetch settings', e);
    } finally {
      this.loading = false;
    }
  }

  async update(data: {
    fileNamingPattern?: string;
    writeMetadataToFile?: boolean;
    maxUploadSizeMb?: number;
  }): Promise<boolean> {
    try {
      this.settings = await apiFetch('/api/settings', {
        method: 'PUT',
        body: JSON.stringify(data)
      });
      return true;
    } catch (e) {
      console.error('Failed to update settings', e);
      return false;
    }
  }
}

export const settingsState = new SettingsState();
