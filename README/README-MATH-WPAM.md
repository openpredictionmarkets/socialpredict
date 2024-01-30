### Weighted Probability Adjustment Model (WPAM)

#### Formula for Updating Market Probability - Weighted Probability Adjustment Model (WPAM)

* The Weighted Probability Adjustment Model is our base model and the math functions for this can be viewed by searching our codebase for WPAM. WPAM is designed to update the probability of an outcome based on the total amount bet on each possibility. It gives more weight to the initial settings (initial probability and investment) to stabilize the market in its early phase.

---
##### WPAM Formula for Updating Market Probability

```math
\begin{align*}
P_{\text{new}} &= \frac{P_{\text{initial}} \times I_{\text{initial}} + A_{\text{YES}}}{I_{\text{initial}} + A_{\text{YES}} + A_{\text{NO}}} \\
\end{align*}
```

```math
\begin{flalign*}
& \text{where:} & \\
& P_{\text{new}} \text{ is the new probability.} & \\
& P_{\text{initial}} \text{ is the initial probability, set to 0.5.} & \\
& I_{\text{initial}} \text{ is the initial investment, assumed to be 10 points.} & \\
& A_{\text{YES}} \text{ is the total amount bet on "YES".} & \\
& A_{\text{NO}} \text{ is the total amount bet on "NO".} &
\end{flalign*}
```

---

* Initial Probability as a Weighted Factor: The initial probability (P_initial) is typically set to represent a balanced or neutral starting point for the market, often 0.5 for a 50-50 chance. This value is used as a weighted factor in the numerator to establish the baseline influence on the market's direction.
* The cost of creating the market is a way of weighting the market as an initial investment (I_initial). While the (P_initial) is meant to represent blind uncertainty, a 50-50 chance of any market created, the (I_initial) is meant to represent a form of stability, which is why it is included in both the numerator and denominator. If there is a larger initial investment, such that (I_initial) >> (A_YES) or (A_NO)  this implies that the market will not move as much until larger (A_YES) or (A_NO) is invested.

#### Example Orders and Outcomes

1. (I_initial) of 10, (P_initial) of 0.50, (A_NO) order made in amount of 20.

```math
P_{\text{new}} = \frac{0.5 \times 10}{10 + 0 + 20} = \frac{5}{30} \approx 0.167
```

2. Same as above but (A_YES) order made in amount of 10.

```math
P_{\text{new}} = \frac{0.5 \times 10 + 10}{10 + 10 + 0} = \frac{15}{20} = 0.75
```

3. Follow up order on (2) made in (A_NO) direction.

```math
P_{\text{new}} = \frac{0.5 \times 10}{10 + 0 + 20} = \frac{5}{30} \approx 0.167
```

#### Market Outcome Update Formulae - Divergence-Based Payout Model (DBPM)

Once a market has been resolved, there is a series of steps that need to be carried out in order to fairly pay out all of the participants based upon how they had bet throughout the duration of the market.

One method of paying out all of the participants might be to simply proportionally cut up the winning pool(s) into shares that are proportional to the amount bet by every participant. However, this doesn't reward those who identified inefficiencies in the market because it doesn't offer much incentive to bet on markets that are either high probability or low probability.

So instead, we need to come up with an operation where every user's bet will be rewarded in proportion to how far it was to the final probability. The following is an explination of the steps we use to calculate the payouts using DBPM.

##### Step One - Dividing Total Payout Pool into YES and NO Payout Pools

* Markets should hypothetically be able to resolve at any given probability. That being said, a complete, "YES" resolution could be defined as resolving at 1 while a complete, "NO" resolution could be defined as 0. Anything in between those numbers could be defined as R.
* If we accept the total pool of bets into the market from the start, meaning the sum of all bet amounts as the total betting tool, then we could calculate the share of that pool, S for either the YES or NO direction.

---
##### DBPM Formula for Dividing Up Market Pool Shares

```math
\begin{flalign*}
& \text{Given:} & \\
& R: \text{Resolution probability (ranging from 0 to 1)} & \\
& S: \text{Total share pool, sum of all bet amounts} & \\
& b_i : \text{Bet amount of bet } i  & \\
& p_i : \text{Market probability at the time of bet } i  & \\
\\
& \text{Total Payout Pools:} & \\
\end{flalign*}
```

```math
\begin{align*}
& S_{\text{YES}} &= S \times R \\
\end{align*}
```

```math
\begin{align*}
& S_{\text{NO}} &= S \times (1 - R) \\
\end{align*}
```
---

##### Step Two - Individual Payout Calculation

* Understanding that we have a series of bets with individual amounts in both YES and NO direction, but we want to incentivize bets that are further away from the R to reward predictors:
* We can introduce d_i, the Reward Factor, representing the linear deviation of the bet's probability p from the resolution probability, R.
* This may be multiplied by the bet amount b_i for any given bet to give us an individual course payout C.

##### DBPM Formula for Calculating Reward Factor

---

```math
\text{For each bet } i: \\
$ d_i = |R - p_i| $ \quad  \\
$ C_i = d_i \times b_i $ \\
```

```math
\begin{flalign*}
& \text{where:} & \\
& R: \text{Resolution probability (ranging from 0 to 1)} & \\
& b_i : \text{Bet amount of bet } i  & \\
& p_i : \text{Market probability at the time of bet } i  & \\
& d_i: \text{the Reward Factor, representing the linear deviation from the R to p_i at b_i} & \\
& C_i: \text{he Course payout prior to normalization} &
\end{flalign*}
```

---

##### Step Three


```math
\text{Step 1: Calculate Raw Payouts} \\
$ C_i = d_i \times b_i \\
\text{where } d_i = |R - p_i| \\

\text{Step 2: Sum Course Payouts for Each Pool} \\
$ C_{\text{YES}} = \sum_{i \in \text{YES}} \text{Raw Payout}_i \\
C_{\text{NO}} = \sum_{i \in \text{NO}} \text{Raw Payout}_i \\

\text{Step 3: Calculate Normalization Factor} \\
$ \text{Normalization Factor}_{\text{YES}} = \min\left(1, \frac{S_{\text{YES}}}{\text{Total Raw Payout}_{\text{YES}}}\right) \\
\text{Normalization Factor}_{\text{NO}} = \min\left(1, \frac{S_{\text{NO}}}{\text{Total Raw Payout}_{\text{NO}}}\right) \\

\text{Step 4: Apply Normalization to Calculate Final Payouts} \\
$ \text{Final Payout}_i = \text{Raw Payout}_i \times \text{Normalization Factor}_{\text{YES or NO}}
```

##### Step Four

```math
\section{Total Payout Distribution}
\
\text{Sum of } $\text{Payout}_i$ \text{ for all YES bets to distribute } $S_{\text{YES}}$ \\
\text{Sum of } $\text{Payout}_i$ \text{ for all NO bets to distribute } $S_{\text{NO}}$
```