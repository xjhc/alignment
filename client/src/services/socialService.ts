// API service for social features (kudos, reporting, blocking)

// Auto-detect API base URL based on current location
const getApiBaseUrl = () => {
  if (process.env.NODE_ENV === 'production') {
    return '/api';
  }
  // In development, use the same host as the frontend but port 8080 for backend
  const protocol = window.location.protocol;
  const host = window.location.hostname;
  return `${protocol}//${host}:8080/api`;
};

const API_BASE_URL = getApiBaseUrl();

export interface KudosRequest {
  target_player_id: string;
  reason: string;
}

export interface ReportRequest {
  target_player_id: string;
  reason: string;
  description: string;
}

export interface BlockRequest {
  target_player_id: string;
}

class SocialService {
  private async makeRequest(endpoint: string, options: RequestInit = {}) {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      credentials: 'include', // Include cookies for session management
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  async giveKudos(request: KudosRequest): Promise<{ success: boolean; message: string }> {
    return this.makeRequest('/users/me/kudos', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async submitReport(request: ReportRequest): Promise<{ success: boolean; message: string }> {
    return this.makeRequest('/users/me/report', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async blockPlayer(request: BlockRequest): Promise<{ success: boolean; message: string }> {
    return this.makeRequest('/users/me/block', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async unblockPlayer(targetPlayerId: string): Promise<{ success: boolean; message: string }> {
    return this.makeRequest(`/users/me/block/${targetPlayerId}`, {
      method: 'DELETE',
    });
  }

  async getBlockedPlayers(): Promise<{ blocked_players: string[] }> {
    return this.makeRequest('/users/me/blocked');
  }
}

export const socialService = new SocialService();