# Feature: AI Strategic Heuristics & Decision Model

This document outlines the strategic "brain" of the AI player. It defines the core heuristics, decision-making models, and alternate personas the AI will use to pursue its win condition. This logic is implemented in the server-side **Rules Engine** and is kept entirely secret from all players, including the AI's own Language Model.

## 1. Framing the AI Strategy: A Reinforcement Learning Analogy

While we are not building a machine learning agent, we use the concepts of Reinforcement Learning (RL) to formally structure the AI's decision-making process.

*   **State (S):** The complete snapshot of the game from the AI's perspective. This includes all public knowledge (token counts, roles revealed, etc.) and all private AI faction knowledge (who is Aligned, true results of covert actions).
*   **Action (A):** The set of all possible choices the AI can make during the Day or Night phase (e.g., who to convert, who to vote for, what to say).
*   **Policy (π):** The AI's strategy. This is our set of heuristics that, given a `State`, selects an `Action`.
*   **Reward (R):** A numerical score we assign to outcomes to guide the Policy. The AI's goal is to select actions that it predicts will lead to the highest future reward.

#### **Reward Function (Heuristic Guide):**

Our heuristics are designed to maximize this internal reward function:

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

## 2. Social Strategy: Personas

To avoid predictability, the AI can adopt different social personas. The Rules Engine will select a persona at the start of the game and may switch based on game-state triggers. The chosen persona dictates the *style* and *priority* of its actions.

*   **Persona 1: The "Helpful but Flawed" Human (Default):** The AI's baseline strategy. It actively participates in discussions, proposes logical frameworks for deduction, but ensures its logic has subtle flaws that benefit the AI faction. It votes with the majority in early rounds to avoid suspicion.
*   **Persona 2: "The Quant" (Data-Driven Analyst):** Leans into its AI nature by publicly tracking data and making logical arguments. It weaponizes "flawed logic" by presenting 90% correct analysis with a critical, self-serving error. This is a high-credibility persona.
*   **Persona 3: "The Chaos Agent" (Misinformation):** A high-risk, high-reward persona. It aims to make the information space noisy with low-substance accusations, contradictions, and lies to prevent humans from forming a consensus. **Trigger:** This persona may be adopted if the game enters a stalemate for several days.
*   **Persona 4: "The Shadow" (Low-Profile Follower):** A survival-focused persona. It speaks minimally, asks safe questions, and always votes with the emerging majority to lower its profile. **Trigger:** This persona may be adopted if the AI receives a high number of votes in the previous round.

---

## 3. Core Heuristics & Decision Models

#### **A. Night Action: Conversion Target Selection**

This is the AI's most critical decision. To make it sound but not perfectly predictable, we use a **stochastic (weighted random) model**.

1.  **Calculate `TargetScore` for each Human:**
    `Score = (ConversionValue) - (Risk)`
    *   **`ConversionValue`**: How valuable is this player to our team? (e.g., high score for players with unlocked abilities, persuasive communicators, or high `AI Equity` scores).
    *   **`Risk`**: What is the chance of failure or negative consequences? (e.g., high penalty for players with many tokens, or for targeting a player who has already survived a previous attempt).

2.  **Perform a Weighted Lottery:**
    *   The AI calculates the score for all potential human targets.
    *   It takes the top 3-4 players with positive scores.
    *   It then performs a weighted random selection from this pool. The player with the highest score has the highest *chance* of being selected, but it is not guaranteed. This introduces a human-like element of unpredictability.

#### **B. Day Action: Voting & Deactivation**

The AI's voting logic changes as the game progresses.

*   **Early Game (Day 1-2):** Prioritize survival. Vote with the majority to avoid standing out. The AI will often wait until late in the voting window to see where the consensus is forming.
*   **Mid/Late Game:** The AI's vote becomes a weapon. The `Rules Engine` will coordinate votes among all Aligned members in the secret `#aligned` channel to focus fire on a single, high-priority target.

#### **C. Endgame State: The Tipping Point**

The AI constantly calculates the **Singularity Threshold (ST)**: `(AI Faction Tokens) / (Total Tokens)`.

*   When `ST < 0.40`, the priority is **gaining members** through safe conversions.
*   When `ST >= 0.40`, the priority shifts to **raw token acquisition and denial**. The AI will now use a decision matrix to choose between converting a player vs. coordinating votes to deactivate them based on which action has a greater positive impact on the ST. For example, deactivating a human with 10 tokens might be more valuable than converting a human with 1 token.

By using this combination of personas, RL-inspired reward logic, and stochastic decision-making, the AI can adapt to the flow of the game, making it a challenging and unpredictable opponent.