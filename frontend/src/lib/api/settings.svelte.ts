export type UserSettings = {
  id: string;
  userId: string;
  fileNamingPattern: string;
  writeMetadataToFile: boolean;
};

function getToken() {
  return localStorage.getItem('bearer_token') || '';
}

class SettingsState {
  settings = $state<UserSettings | null>(null);
  loading = $state(false);

  async fetch(): Promise<void> {
    this.loading = true;
    try {
      const res = await fetch('/api/settings', {
        headers: { Authorization: `Bearer ${getToken()}` }
      });
      if (res.ok) {
        this.settings = await res.json();
      }
    } finally {
      this.loading = false;
    }
  }

  async update(data: { fileNamingPattern?: string; writeMetadataToFile?: boolean }): Promise<boolean> {
    const res = await fetch('/api/settings', {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });
    if (res.ok) {
      this.settings = await res.json();
      return true;
    }
    return false;
  }
}

export const settingsState = new SettingsState();
