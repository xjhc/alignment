# Legal: Data Privacy, ToS, and Technical Compliance

## 1. Purpose

This document outlines our strategy for handling user data, defines our terms of service, and maps these legal requirements to our technical architecture. It serves as an internal guide for developers to ensure our application is built in compliance with our stated policies.

**Core Principle:** We will be transparent with our users, collect only the data we absolutely need to operate the service, and protect that data responsibly.

## 2. Key Legal Documents

The following are the official, user-facing legal documents for `Alignment`. They will be hosted on our main website and linked from within the game client.

*   **[Terms of Service (ToS)](./02-terms-of-service.md):** The contract between us and our players. It defines the rules for using our service.
*   **[Privacy Policy](./03-privacy-policy.md):** Explains what data we collect, why we collect it, and how we use and protect it.
*   **[Community Guidelines](./04-community-guidelines.md):** A user-friendly version of the moderation rules from our ToS.

These documents will be drafted and reviewed by legal counsel before launch.

## 3. Data Privacy & Technical Implementation

This section details how our architecture fulfills the promises made in our Privacy Policy.

#### **A. Data We Collect**

We must be precise about what we store.

*   **Account Information (PII - Personally Identifiable Information):**
    *   `Player Handle/Username`: Publicly visible.
    *   `Authentication ID`: A unique ID from a third-party auth provider (e.g., Discord, Google) or a hashed password if we implement our own auth. **We never store raw passwords.**
    *   `Email Address`: Collected for account recovery and essential notifications. Stored securely and not shared.
    *   `Creation Date`: Timestamp of account creation.
*   **Game-Generated Data (Non-PII):**
    *   `Game Logs`: The event stream for each game, stored in Redis. Contains `player_id`s and chat messages.
    *   `Player Stats`: Aggregated, anonymous statistics for meta-progression (wins, losses, achievements).
    *   `User Reports`: Reports filed by players against others, including chat logs for context.
*   **Operational Data (Non-PII):**
    *   `Server Logs`: Application logs containing IP addresses for security and debugging purposes.

#### **B. Data Retention Policy & Implementation**

Our goal is to minimize the data we hold.

*   **Requirement:** Game-specific data (chat logs, event streams) should be automatically deleted after a set period.
*   **Implementation:**
    *   **Redis:** All game-related keys (`game:*:events`, `game:*:snapshot`) will be created with a **7-day TTL (Time To Live)**. Redis will automatically purge this data. This is already noted in `server/README.md`.
    *   **Server Logs:** Logs containing IP addresses will be stored for a maximum of **30 days** for security analysis and then purged. This will be managed by our logging service configuration.
*   **Requirement:** Player accounts and meta-progression data are kept as long as the account is active.
*   **Implementation:** We will provide a clear "Delete Account" button in the user's profile settings. This action will trigger a process to permanently delete their account information and anonymize their associated stats.

## 4. Operational Requirements for Legal Compliance

This checklist translates our legal documents into concrete operational tasks for the engineering team.

*   **[ ] Cookie Consent Banner:** Implement a cookie consent banner on our website and landing page, especially if we use analytics tools like Google Analytics.
*   **[ ] "I Agree" Checkbox:** During account creation, there must be a mandatory, unticked checkbox: `[ ] I have read and agree to the Terms of Service and Privacy Policy.` with links to both documents. A player cannot create an account without checking this box.
*   **[ ] In-Game Links:** The game's settings menu must contain persistent links to the ToS and Privacy Policy.
*   **[ ] Account Deletion Flow:**
    *   **UI:** An easily accessible "Delete Account" button within the user's profile settings.
    *   **Confirmation:** This action must trigger a confirmation modal that clearly states: "This action is irreversible and will permanently delete your account, stats, and all associated data. Are you sure you wish to proceed?"
    *   **Backend:** A secure API endpoint (`DELETE /api/users/me`) that triggers a background job to scrub the user's PII from the database.
*   **[ ] Data Portability (Future Consideration for GDPR):** While not required for V1, our architecture should be able to support a "Download My Data" feature in the future. Since our data is structured, this is feasible. We will defer implementation until required.