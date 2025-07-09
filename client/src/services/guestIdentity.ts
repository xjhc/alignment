import { v4 as uuidv4 } from "uuid";
import type { UserIdentity } from "../types";

const GUEST_ID_KEY = "alignment_guest_id";
const GUEST_PROFILE_KEY = "alignment_guest_profile";

export interface GuestProfile {
  id: string;
  name: string;
  avatar: string;
  createdAt: string;
  gamesPlayed: number;
}

/**
 * Gets or creates a guest ID, storing it in localStorage for persistence
 */
export function getOrCreateGuestId(): string {
  let guestId = localStorage.getItem(GUEST_ID_KEY);

  if (!guestId) {
    guestId = `guest:${uuidv4()}`;
    localStorage.setItem(GUEST_ID_KEY, guestId);
  }

  return guestId;
}

/**
 * Gets the current guest profile or creates a default one
 */
export function getGuestProfile(): GuestProfile | null {
  const profileData = localStorage.getItem(GUEST_PROFILE_KEY);
  if (!profileData) {
    return null;
  }

  try {
    return JSON.parse(profileData);
  } catch {
    // Invalid profile data, remove it
    localStorage.removeItem(GUEST_PROFILE_KEY);
    return null;
  }
}

/**
 * Updates the guest profile with new information
 */
export function updateGuestProfile(
  profile: Partial<Omit<GuestProfile, "id">>
): GuestProfile {
  const guestId = getOrCreateGuestId();
  const existingProfile = getGuestProfile();

  const updatedProfile: GuestProfile = {
    id: guestId,
    // If a name is provided, use it. Otherwise, generate a guest name.
    // This enforces that guest logins (which won't provide a name) get a generated name.
    name: profile.name || `Guest#${guestId.substring(6, 10)}`,
    avatar: profile.avatar || existingProfile?.avatar || "👤",
    createdAt: existingProfile?.createdAt || new Date().toISOString(),
    gamesPlayed: profile.gamesPlayed ?? existingProfile?.gamesPlayed ?? 0,
  };

  localStorage.setItem(GUEST_PROFILE_KEY, JSON.stringify(updatedProfile));
  return updatedProfile;
}

/**
 * Clears guest identity data (used when user authenticates)
 */
export function clearGuestIdentity(): void {
  localStorage.removeItem(GUEST_ID_KEY);
  localStorage.removeItem(GUEST_PROFILE_KEY);
}

/**
 * Gets the current user identity (guest or authenticated)
 */
export function getCurrentUserIdentity(): UserIdentity | null {
  // TODO: Check for authenticated user first when auth is implemented
  const guestProfile = getGuestProfile();

  if (guestProfile && guestProfile.name) {
    return {
      id: guestProfile.id,
      name: guestProfile.name,
      avatar: guestProfile.avatar,
      isAuthenticated: false,
    };
  }

  return null;
}

/**
 * Gets the user ID to send to the server (guest ID or authenticated user ID)
 */
export function getUserIdForApi(): string {
  const userIdentity = getCurrentUserIdentity();
  if (userIdentity) {
    return userIdentity.id;
  }

  // Fallback to just the guest ID if no profile exists yet
  return getOrCreateGuestId();
}
