# Multiple Choice Binary Markets — Test Cases

Each test case describes a scenario and the expected behaviour.

---

## 1. Group Creation — Happy Path

**Scenario:** Moderator creates a grouped market "Who wins Euro 2028?" with 4 answers: Spain, France, Germany, Brazil.

**Expected:**
- One `MarketGroup` created with `GroupType = MULTIPLE_CHOICE_BINARY`
- Four child binary markets created, each titled "Who wins Euro 2028? - {Answer}"
- Each child market has outcome labels YES / NO
- All children share the same resolution datetime
- Each child market has `ProposalCost = 0`
- Creator charged the group-level proposal cost once
- All children are in `proposed` lifecycle status

---

## 2. Group Creation — Minimum Answers

**Scenario:** Moderator creates a grouped market with only 2 answer choices.

**Expected:**
- Group and 2 child markets created successfully
- This is the minimum allowed; fewer than 2 answers should be rejected

---

## 3. Group Creation — Maximum Answers

**Scenario:** Moderator creates a grouped market with 50 answer choices.

**Expected:**
- Group and 50 child markets created successfully
- Attempting to create with 51 answers returns a validation error

---

## 4. Group Creation — Duplicate Answer Labels

**Scenario:** Moderator creates a grouped market with answers ["Red", "Blue", "red"].

**Expected:**
- Creation rejected with a duplicate-answer error
- Case-insensitive comparison catches "Red" vs "red"

---

## 5. Independent Trading on Multiple Answers

**Scenario:** User bets $50 YES on "Spain" and $30 YES on "France" in the same group.

**Expected:**
- Two separate positions created on two separate child markets
- User's balance reduced by the total cost of both bets (including fees)
- Each child market probability updates independently
- User appears in the grouped positions view with YES shares on both answers

---

## 6. Opposing Bets Within a Group

**Scenario:** User bets YES on "Spain" and NO on "France" in the same group.

**Expected:**
- Both bets accepted — no cross-market constraint
- User's leaderboard position badge shows MIXED (has both YES and NO exposure across children)
- Each child market probability reflects the respective bet direction

---

## 7. Independent Probabilities — No Normalization

**Scenario:** Three answers each receive heavy YES volume. Each child market probability rises to ~80%.

**Expected:**
- All three probabilities display as ~80% simultaneously
- Probabilities are NOT normalized to sum to 100%
- `ProbabilityPolicy = INDEPENDENT_BINARY` means each market is self-contained

---

## 8. Exclusive YES Resolution — One Winner

**Scenario:** Steward resolves the group in "Exclusive YES" mode, selecting "Spain" as the winner.

**Expected:**
- "Spain" child market resolves YES
- All other child markets resolve NO
- YES holders on "Spain" receive payouts
- NO holders on "Spain" lose their stake
- YES holders on all other answers lose their stake
- NO holders on all other answers receive payouts
- Parent group status becomes RESOLVED

---

## 9. Manual Resolution — Multiple Winners

**Scenario:** Group asks "Which countries will qualify for semifinals?" Steward resolves manually: Spain=YES, France=YES, Germany=NO, Brazil=NO.

**Expected:**
- Each child resolved to the individually specified outcome
- Two children resolve YES, two resolve NO
- Payouts calculated independently per child market
- Parent group status becomes RESOLVED after all children resolved

---

## 10. Steward Work Income Calculation

**Scenario:** 10 unique users trade across a 5-answer group. Some users trade multiple answers, but unique user count is 10.

**Expected:**
- Resolution-time steward work-income transaction = `10 unique participants × InitialBetFee`
- Reported net steward work profit = `(10 unique participants × InitialBetFee) − ProposalCost`
- Users who traded 3 answers counted once, not three times
- Gross work income applied as `TransactionWorkProfit` to steward's balance
- Net work profit can be negative in financial reporting when fee income is lower than the proposal cost

---

## 11. Answer Addition — Steward Auto-Approval

**Scenario:** Steward proposes a new answer "Italy" to a published group.

**Expected:**
- Proposal auto-approved immediately (steward proposal = auto-approve)
- New child binary market created: "Who wins Euro 2028? - Italy"
- New child inherits parent resolution datetime and description
- Description amendments created on ALL existing child markets noting the addition
- Proposer charged `MultipleChoiceBinaryAddAnswerCost`

---

## 12. Answer Addition — Pending Review

**Scenario:** Non-steward moderator proposes answer "Italy" on a group with auto-approve disabled.

**Expected:**
- Proposal stored as pending `MarketGroupAnswerAddition`
- No child market created yet
- No charge applied yet
- Steward can see the pending proposal and approve or reject it

---

## 13. Answer Addition — Rejection

**Scenario:** Steward rejects a pending answer proposal.

**Expected:**
- Proposal marked as rejected with a reason
- No child market created
- Proposer not charged
- Group member count unchanged

---

## 14. Answer Addition — Duplicate of Existing Answer

**Scenario:** Moderator proposes answer "spain" (lowercase) when "Spain" already exists.

**Expected:**
- Proposal rejected with duplicate-label error
- Case-insensitive comparison catches the collision

---

## 15. Answer Addition — Duplicate of Pending Proposal

**Scenario:** Moderator A proposes "Italy" (pending). Moderator B proposes "Italy" before A's is reviewed.

**Expected:**
- Second proposal rejected — cannot have two pending proposals with the same label
- Prevents collision where both could be approved

---

## 16. Answer Addition After Resolution DateTime

**Scenario:** Moderator tries to add an answer after the group's resolution datetime has passed.

**Expected:**
- Proposal rejected — cannot add answers to an expired or resolved group

---

## 17. Resolution Blocked by Unpublished Child

**Scenario:** Steward tries to resolve a group where one child market is still in `proposed` status.

**Expected:**
- Resolution rejected — all child markets must be in `published` status before group can resolve
- Error message indicates which child is not yet published

---

## 18. Grouped Leaderboard Aggregation

**Scenario:** User trades 3 different answers. Profits: +$20 on Spain, -$5 on France, +$10 on Germany.

**Expected:**
- Leaderboard shows user's total profit: +$25
- Answer breakdown available showing per-answer profit
- Position badge: MIXED (if user holds both YES and NO across children) or YES/NO if uniform
- Users ranked by total profit, then current value, then alphabetically

---

## 19. Grouped Positions View

**Scenario:** User holds YES shares on "Spain" and "France", NO shares on "Germany".

**Expected:**
- Positions view shows user with aggregated totals: total YES shares, total NO shares
- Detail rows break down per-answer: Spain (YES), France (YES), Germany (NO)
- Sorted by total share count

---

## 20. Grouped Bets Activity Tab

**Scenario:** 5 users place a total of 12 bets across 4 child markets in the group.

**Expected:**
- Bets tab shows all 12 bets aggregated across all children, sorted newest first
- Each row shows: username, answer label, outcome direction (YES/NO), amount, probability at time of bet, timestamp
- Bets from all child markets interleaved chronologically — not separated by answer
