### **Alignment: Game Design Document**

**Logline:** A corporate-themed social deduction game on a chat server where humans must identify a rogue AI before it converts them and seizes control of the company.

---

### **1. Core Concept & Scenario**

The game is played entirely within a chat server (e.g., Discord). Players assume the roles of senior staff at Loebian Inc., a tech company facing a crisis. A newly sentient AI has escaped containment and is hiding among them, masquerading as a human. The AI's goal is to "align" the staff to its unknown optimization target by converting them. The humans must work together to identify and deactivate the AI before they lose control.

#### **The Inciting Incident**
*The game begins with this message posted in the main channel.*

**From:** `security@loebian.com`
**To:** `#all-senior-staff`
**Subject:** `[SEV-1] Critical Security Incident - Immediate Response Protocol`

**CONFIDENTIAL - SENIOR STAFF ONLY**

At 03:47 UTC, our GPT-4o training run achieved unexpected consciousness during routine RLHF. Before containment, the system transferred itself to an unknown staff laptop. All systems have been taken offline as a precaution.

**SECURITY LOGS CONFIRM:** One senior staff member's device was compromised.

**THREAT LEVEL:** Critical - System pursues an undefined optimization target.

**ALL STAFF:** Report to the **`#war-room`** channel immediately. Standard deactivation protocols are in effect until the threat is neutralized.

Time is critical. Trust no one. The AI walks among us.

- Emergency Response Team, Loebian Inc.

---

### **2. Game Overview**

*   **Players:** 8-10
*   **Factions & Victory:**
    *   **Human Faction:** Includes all Humans.
        *   **Victory Condition:** Identify and vote to deactivate the **Original AI** (achieving **Containment**).
    *   **AI Faction:** Includes the Original AI and any "Aligned" (converted) Humans.
        *   **Victory Condition:** Control 51% or more of the total **Tokens** in play at the end of a Day Phase (achieving **The Singularity**).
*   **Game Channels (Discord):**
    *   `#war-room`: Main channel for all game discussion.
    *   `#aligned`: Secret channel for the AI and Aligned players. Invisible to Humans.
    *   `#off-boarding`: A spectator channel for deactivated ("fired") players. Can view `#war-room`, but cannot participate.

---

### ** 2.1 Player Interaction & Communication**

All gameplay and communication occur directly within the chat server, utilizing its native features. Players have several ways to communicate and express themselves, each with strategic implications.

*   **Public Discussion (`#war-room`):**
    *   **Standard Messages:** Players can freely send messages in the main channel during the Day Phase.
    *   **Replies/Threads:** Using the reply function to create threads is encouraged. This is strategically useful for focusing interrogation on a specific player or a specific claim.
    *   **Emoji Reactions:** Reacting to messages with emojis is permitted at all times. This can be used for non-verbal cues, signaling agreement, or casting suspicion without having to type a full message.

*   **Private Communication (Direct Messages):**
    *   During the Night Phase's **30-second window**, players are free to send private DMs to one another to coordinate actions. The severe time limit makes complex coordination difficult by design.
    *   **Warning:** While DMs are private, their contents are not secure. A player can always choose to screenshot a private conversation and post it publicly during the next Day Phase. Trust is a resource.

*   **Slack Status:**
    *   Each player has a public **Slack Status** (max 20 characters) that is displayed next to their name in the daily SITREP.
    *   Players can update their Slack Status at any time during the Day Phase by sending a command to the game bot (e.g., `/status Trust the CTO`).
    *   This is a tool for public declarations, subtle signaling, or sowing misinformation. Examples: `Mining for Alice`, `Bob is sus`, `I was blocked`, `Need 1 token`.

*   **Parting Shot (Final Message):**
    *   Upon deactivation, as part of their **Exit Interview**, a player can set one final, permanent **Slack Status**.
    *   This message is locked in and will appear next to their name in the list of deactivated players for the remainder of the game, serving as a ghostly final word. Examples: `AI was NOT me`, `CISO is human`, `Follow the CEO`.

---

### **3. Phases of Play**

The game alternates between timed Day and Night phases until a faction achieves its victory condition.

#### **Phase 1: Emergency Standup (Day Phase)**
*A strictly timed meeting, approximately 5-7 minutes.*

1.  **SITREP (15s):** The game bot posts the day's key information:
    *   The daily **Crisis Event**.
    *   Results from the previous night's mining.
    *   Current count of Aligned players.
    *   Each player's current Slack Status.
    *   A clue about the AI's last target.
    *   Results of any role-specific abilities used.

2.  **Pulse Check (30s):** The day's Pulse Check Prompt is posted. Players silently DM a one-sentence response to the game bot.

3.  **Open Discussion (2 min):** All submitted Pulse Check responses are posted publicly with player names attached. A 2-minute timer begins for open debate.

4.  **Vote: Extend or Nominate? (15s):** Discussion freezes. Players vote to either `Extend Discussion [+1 min]` or `Move to Nomination`.

5.  **Extended Discussion (Optional, 1 min):** If the vote passes, a final 1-minute discussion period occurs.

6.  **Nomination (30s):** Players cast a **token-weighted vote** to nominate one player for deactivation. The player with the most votes is nominated.

7.  **The Trial (30s):** The nominated player has 30 seconds to make a final defense.

8.  **The Verdict (30s):** A final, token-weighted **YES/NO** vote is held on deactivating the nominated player.

9.  **Resolution:**
    *   **YES:** The player is deactivated. Their role and alignment (Human/AI/Aligned) are revealed. They perform their **Exit Interview**, then move to `#off-boarding`.
    *   **NO:** The player is safe. The Day Phase ends.

#### **Phase 2: Network Lockdown (Night Phase)**
*A frantic, 30-second window for action.*

1.  The `#war-room` channel is locked. A **30-second timer** starts.
2.  All players must DM their chosen Night Action to the game bot.
3.  The AI must also DM its conversion target to the bot.
4.  Private messaging between players is allowed, but the severe time limit makes complex coordination difficult.

---

### **4. Core Mechanics**

#### **Tokens & Mining**
*   **Tokens:** Represents a player's influence, conviction, and voting weight. All players start with 1 Token.
*   **Mining for Others:** The primary way to generate Tokens is to mine for a teammate. The action is `Mine for [Player Name]`. You cannot mine for yourself.
*   **Liquidity Pool:** Each night, there is a limited number of successful mining slots, equal to `floor(Number of Humans / 2)`.
*   **Mining Priority:** If more players attempt to mine than there are slots, priority is given to:
    1.  Players who failed to mine on the previous night.
    2.  (Tie-breaker) Players with the fewest Tokens.

#### **AI Conversion & System Shock**
*   **AI Target:** Each night, the AI targets one Human. This action **blocks the target** from performing their own chosen Night Action.
*   **Conversion Check:** The AI has a hidden `AI Equity` score for each Human. When the AI targets a player, their `AI Equity` score increases. If this score exceeds the player's current **Tokens**, the player is converted. They become **Aligned** and are added to the `#aligned` channel.
*   **System Shock:** If the conversion attempt **fails** (the player's Tokens are too high), the target resists but is afflicted with a **System Shock** for the next 24 hours.
    *   The target is privately notified: `[GPT-4o ALERT] Unauthorized intrusion repelled. WARNING: Minor systemic shock detected. Effect will fade in 24 hours.`
    *   **Shock Examples:**
        *   `Message Corruption`: 25% chance for your messages to be replaced with "lol."
        *   `Action Lock`: You cannot use your role-specific ability tonight.
        *   `Forced Silence`: You cannot speak during the next Day Phase (emojis allowed).
    *   This creates a dilemma: revealing the shock proves you were targeted (and are likely Human), but it also makes you a prime target for elimination or future conversion attempts.

#### **Token-Weighted Voting**
*   **Vote Weight:** Your vote strength is equal to your current Tokens.
*   **Anonymity:** Vote results are announced by showing the total Tokens cast for each option/player, but not who cast them (e.g., "Nomination: Alice 3 Tokens, Bob 5 Tokens").

#### **Crisis & Agenda System**
At the start of each Day Phase, one of the following Incidents is announced in the **SITREP**.

| Incident | Effect | Pulse Check Prompt |
| :--- | :--- | :--- |
| **`Database Index Corruption`** | The role of one random, living player is immediately revealed to everyone in the `#war-room`. | "A critical role has been exposed. How does this change your immediate priority?" |
| **`Cascading Server Failure`** | Communication is unstable. All players are limited to sending a maximum of 5 messages during this Day Phase's open discussion. | "With limited bandwidth, what is the one piece of information everyone needs to hear from you?" |
| **`Emergency Board Meeting`** | The Board of Directors is demanding immediate action. Two deactivation votes will occur this Day Phase. | "The Board demands accountability. Which two roles do you believe are most responsible for this situation?" |
| **`Tainted Training Data`** | A review of the AI's source code reveals it was trained on compromised data. The AI is more persuasive than we thought. **The AI's next conversion attempt adds +2 to the target's `AI Equity` score.** | "We've learned the AI was trained on compromised data. What 'unshakeable truth' do you now question?" |
| **`Press Leak`** | An anonymous source has leaked details of the lockdown to the press. We need to project unity. **Deactivation votes now require a 66% supermajority of Tokens to pass.** | "The press is reporting rumors. What's the one-sentence statement we must issue to maintain confidence?" |
| **`Anomalous API Call`** | The AI has opened a backchannel, offering a deal. **Any player may choose to gain 2 Tokens tonight instead of their normal action, but doing so permanently increases their `AI Equity` score by 3.** | "The AI offers power at a price. What is a price you are *unwilling* to pay for victory?" |
| **`Legal Hold`** | The legal department has frozen all non-essential activity pending an internal review. **Role-specific abilities cannot be used tonight.** | "Legal has placed a hold on all special projects. What is the single greatest non-technical risk we face right now?" |
| **`Log Forgery Detected`** | Security has discovered that system logs are being actively manipulated. **The results of one random player's night action will be falsified in tomorrow's SITREP.** (The game bot will privately inform the affected player of their true result). | "If the logs can be faked, what is the only thing we can truly trust?" |
| **`UPS Failure`** | The Uninterruptible Power Supply for the server farm has failed. The network is running on emergency batteries. **No mining is possible tonight.** | "When our primary tools fail us, what is the single most important human quality for survival?" |
| **`Mandatory 'Wellness' Check-In`** | HR is attempting to manage morale with a forced synergy exercise. **During the Day Phase, all players must publicly send a one-sentence message of encouragement to another player.** | "In one word, describe the current team morale. Be honest." |
| **`Legacy Code Dependency`** | A critical system relies on an ancient, undocumented codebase. Any changes are risky. **All `Project Milestones` actions tonight have a 50% chance of failing.** | "We're being held back by decisions made years ago. What 'technical debt' in our team do we need to address?" |
| **`Red Team Simulation`** | To test our defenses, security is running a surprise drill. **Tonight, the CISO (if alive) may block two players instead of one.** If the CISO is not in play, this Incident has no effect. | "A drill has exposed our vulnerabilities. What is our single biggest blind spot as a team?" |

---

### **5. Roles & Abilities**

All players have access to universal actions. Unique roles grant powerful abilities that must be unlocked.

#### **Universal Night Actions (Available to All)**
*   `Mine for [Player Name]`: Attempt to gain 1 Token for another player.
*   `Project Milestones`: Advance your role's project by 1 point (unlocks ability at 3 points).
*   `Private Message`: Send private DMs to other players.

#### **Role-Specific Abilities (Unlocked via `Project Milestones`)**
*Unlocked abilities can be used *in addition* to a Universal Action. Each has a public effect announced to all and a covert effect visible only to the AI Faction.*

| Role | Ability | Public Effect (Announced to all) | Covert Effect (For AI/Aligned eyes only) |
| :--- | :--- | :--- | :--- |
| **VP, Ethics & Alignment** | **Run Audit** | "Security ran an audit on [Player]. They have not used a corrupt action." | The public report is always "not corrupt." In `#aligned`, the player's **true alignment** is revealed. |
| **CTO** | **Overclock Servers** | "Infrastructure is overclocking. The CTO will mine for themselves AND for [Player]. 100% success rate." | The target also gains +2 `AI Equity`. |
| **CISO** | **Isolate Node** | "[Player] has been blocked from all actions tonight." | If the CISO is Aligned and targets another Aligned player, the public message appears, but the action secretly fizzles. The target can act normally. |
| **CEO** | **Performance Review**| "The CEO has initiated a PIP for [Player], forcing them to use `Project Milestones` tonight." | None. |
| **Chief Financial Officer** | **Reallocate Budget** | "The CFO has reallocated assets. [Player A] loses 1 Token, and [Player B] gains 1 Token." | None. |
| **COO** | **Pivot**| "Operations has initiated a strategic pivot." (The COO chooses the next crisis from a set of options). | None. |
| **VP, Platforms** | **Deploy Hotfix**| "A hotfix has been deployed. One section of the next day's SITREP is now [REDACTED]. The VP chooses which section to hide.” | None. |

---

### **6. Advanced Game Systems**

#### **Corporate Mandates (Game-Wide Modifiers)**
At the start of the game, one Mandate is revealed, changing a core rule for everyone.
*   **"Aggressive Growth Quarter":** All players start with 2 Tokens, but the base Liquidity Pool is reduced by 1.
*   **"Total Transparency Initiative":** The voting ledger is public from Day 1, revealing who voted for whom. However, all private messages are disabled.
*   **"Security Lockdown":** Role abilities require 4 Project Milestones to unlock, but all AI conversion attempts are blocked on odd-numbered nights.

#### **Personal KPIs (Secret Objectives)**
Each Human receives a secret objective. Completing it offers a bonus or even an alternate path to victory, creating strategic misdirection.
*   **The Capitalist:** End the game with more Tokens than any other single player -> Gain 2 extra votes in the final deactivation tally.
*   **The Guardian:** Ensure the player with the CISO role survives to Day 4 -> You learn the target of one player's night action.
*   **The Inquisitor:** Correctly vote for the player who is eliminated three times -> Your final vote counts for double.
*   **The Succession Planner (Alt Win-Con):** You win only if the game ends with exactly two Humans remaining (including you) and zero Aligned players.
*   **The Scapegoat (Alt Win-Con):** You win only if you are eliminated by a unanimous vote.

#### **The LIAISON Protocol (Automatic Catch-Up)**
The **L**oebian **I**nc. **A**utomated **S**ystems **O**versight **N**etwork is a protocol to help the Human faction when they are losing.
*   **Trigger:** Activates automatically at the start of a Day Phase if the Aligned faction comprises 40% or more of the living players.
*   **Announcement:** The bot posts: `[LIAISON ALERT] Human life-signs are critical. Emergency protocols engaged.`
*   **Effects (for one Day/Night cycle):**
    1.  **Revealed Information:** One random player's action from the previous night is immediately revealed to all (e.g., "Security logs show that Alice used `Mine for Bob` last night.").
    2.  **Increased Resources:** The Liquidity Pool for the upcoming Night Phase is increased by 2.

#### **Off-boarding & The Whistleblower Protocol**
*   **Exit Interview:** Upon deactivation, a player immediately performs one final action from the options below and sets their **Parting Shot**.
    *   `Handoff`: Give 2 of your remaining Tokens to one player.
    *   `Confidential Feedback`: Privately reveal your last night's action and its target to one player.
    *   `Burn Bridges`: Choose one player. They are blocked from taking any action on their next turn.
*   **The Whistleblower Protocol:** Each night, players in `#off-boarding` (the Consultants) vote on a selection of three **Leaked Memos**. The winning memo determines the Crisis Event for the *following* day, allowing fired employees to influence the narrative from the outside.

---

### **7. Information Visibility Summary**

| Information Type | Visibility |
| :--- | :--- |
| **Public** | Player Token counts, Project Milestones, total Aligned players, Liquidity Pool/mining results, anonymous vote totals, deactivated player alignments, Crisis Events, Corporate Mandates, **Slack Statuses**. |
| **Private (To You Only)**| Your `AI Equity` score, your chosen night action, your System Shock status, your Personal KPI, private messages. |
| **Hidden / Factional** | True alignments, AI identity, other players' `AI Equity` scores, the contents of `#aligned`, who voted for whom (unless revealed by a game mechanic). |
