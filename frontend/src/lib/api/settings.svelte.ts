import { apiFetch } from './client';

export type UserSettings = {
  id: string;
  userId: string;
  fileNamingPattern: string;
  writeMetadataToFile: boolean;
  maxUploadSizeMb: number;
};

export type OPDSSettings = {
  username: string;
  enabled: boolean;
};

class SettingsState {
  settings = $state<UserSettings | null>(null);
  opds = $state<OPDSSettings | null>(null);
  loading = $state(false);

  async fetch(): Promise<void> {
    this.loading = true;
    try {
      this.settings = await apiFetch('/api/settings');
      this.opds = await apiFetch('/api/settings/opds');
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

  async updateOPDS(data: {
    username?: string;
    password?: string;
    enabled?: boolean;
  }): Promise<boolean> {
    try {
      await apiFetch('/api/settings/opds', {
        method: 'PUT',
        body: JSON.stringify(data)
      });
      this.opds = await apiFetch('/api/settings/opds');
      return true;
    } catch (e) {
      console.error('Failed to update OPDS settings', e);
      return false;
    }
  }
}

export const settingsState = new SettingsState();
