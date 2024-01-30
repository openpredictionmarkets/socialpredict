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

#### Market Outcome Update Formulae - Weighted Probability Adjustment Model (WPAM)

* Markets should hypothetically be able to resolve at any given probability. That being said, a complete, "YES" resolution could be defined as resolving at 1 while a complete, "NO" resolution could be defined as 0. Anything in between those numbers could be defined as R.
* If we accept the total pool of bets into the market from the start, meaning the sum of all bet amounts as the total betting tool, then we could calculate the share of that pool, S for either the YES or NO direction.

```math
\begin{flalign*}
\text{Given:} \\
R: \text{Resolution probability (ranging from 0 to 1)} \\
S: \text{Total share pool, sum of all bet amounts} \\
$ b_i $: \text{Bet amount of bet } i \\
$ p_i $: \text{Market probability at the time of bet } i \\
\\
\text{Total Payout Pools:} \\
\\
\end{flalign*}
```

```math
\begin{align*}
& S_{\text{YES}} &= S \times R \\
\end{lalign*}
```

```math
\begin{align*}
& S_{\text{NO}} &= S \times (1 - R) \\
\end{lalign*}
```
