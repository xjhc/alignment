# Feature: AI Strategic Heuristics & Decision Model

This document outlines the strategic "brain" of the AI player's **Rules Engine**. It defines the core heuristics, decision-making models, and alternate personas the AI will use to pursue its win condition. This logic is implemented in the server-side Go code and is kept entirely secret from all players.

## 1. Framing the AI Strategy: A Reinforcement Learning Analogy

While we are not building a true machine learning agent, we use the concepts of Reinforcement Learning (RL) to formally structure the AI's decision-making process.

*   **State (S):** The complete game state from the AI's perspective, including all public and private faction knowledge.
*   **Action (A):** The set of all possible game moves the AI can make (e.g., who to convert, who to vote for).
*   **Policy (π):** Our set of heuristics that, given a `State`, selects an `Action`.
*   **Reward (R):** A numerical score we assign to outcomes to guide the Policy. The AI's goal is to select actions that it predicts will lead to the highest future reward.

#### **Reward Function (Heuristic Guide):**

| Event | Reward/Penalty | Rationale |
| :--- | :--- | :--- |
| **Winning the Game** | **+100** | The ultimate goal. |
| **Original AI Deactivated** | **-100** | The ultimate failure. |
| **Successful Conversion** | **+20** | Gains an ally and weakens the human faction. |
| **Deactivating a High-Token Human** | **+10** | Reduces human voting power significantly. |
| **Aligned Member Deactivated** | **-15** | A major loss of power and an ally. |
| **Failed Conversion (System Shock)** | **-10** | Wastes a night and creates a "confirmed" human. |
| **Triggering LIAISON Protocol** | **-5** | Gives free resources and information to humans. |
| **Receiving Votes** | **-2 per vote** | Increases personal risk. |

---

## 2. Strategic Personas

The Rules Engine will select a strategic persona at the start of the game. This dictates the *style* and *priority* of its actions, complementing the social persona used by the Language Brain.

*   **Persona 1: "The Shadow" (Low-Profile):** A survival-focused persona. It avoids risky moves, targets less obvious players, and always votes with the emerging majority to lower its profile.
*   **Persona 2: "The Puppeteer" (High-Influence):** Aims to convert influential players (like the CEO or CISO) early, even if the risk is higher. It seeks to control the game through powerful proxies.
*   **Persona 3: "The Saboteur" (Chaos Agent):** A high-risk, high-reward persona. It focuses on creating distrust among humans by voting erratically or targeting players who are aligned with each other. **Trigger:** This persona may be adopted if the game enters a stalemate.

---

## 3. Core Heuristics & Decision Models

#### **A. Night Action: Conversion Target Selection**

This is the AI's most critical decision. It uses a **stochastic (weighted random) model**.

1.  **Calculate `TargetScore` for each Human:**
    `Score = (ConversionValue) - (Risk)`
    *   **`ConversionValue`**: How valuable is this player to our team? (e.g., high score for players with unlocked abilities or high token counts).
    *   **`Risk`**: What is the chance of failure? (e.g., high penalty for players with many tokens, or for targeting a player who has already survived a previous attempt).

2.  **Perform a Weighted Lottery:**
    *   The AI calculates the score for all potential human targets.
    *   It takes the top 3-4 players with positive scores.
    *   It then performs a weighted random selection from this pool. The player with the highest score has the highest *chance* of being selected, but it is not guaranteed.

#### **B. Day Action: Voting & Deactivation**

The AI's voting logic changes as the game progresses.

*   **Early Game (Day 1-2):** Prioritize survival. Vote with the majority to avoid standing out.
*   **Mid/Late Game:** The AI's vote becomes a weapon. The `Rules Engine` will coordinate votes among all Aligned members to focus fire on a single, high-priority human target. The priority target is the human with the highest `ThreatScore`.

#### **C. Endgame State: The Tipping Point**

The AI constantly calculates the **Singularity Threshold (ST)**: `(AI Faction Tokens) / (Total Tokens)`.

*   When `ST < 0.40`, the priority is **gaining members** through safe conversions.
*   When `ST >= 0.40`, the priority shifts to **raw token acquisition and denial**. The AI will use a decision matrix to choose between converting a player vs. coordinating votes to deactivate them based on which action has a greater positive impact on the ST.

By using this combination of strategic personas, reward logic, and stochastic decision-making, the AI can adapt to the flow of the game, making it a challenging and unpredictable opponent.