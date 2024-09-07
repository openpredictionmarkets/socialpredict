### Market Probability Adjustment and Outcome Update Mechanisms

Two phases of [Double Auctions](https://en.wikipedia.org/wiki/Double_auction) occuring over time arguably exist:

1. The time from the start of the auction to just before the end of the auction, where buyers and sellers are mutually determining a price at any given time during the auction.
2. The end price of the market, which determines how all participants get paid out.

To that end, we have established two mathematical frameworks which allow buying and selling of shares in a binary market to allow (1) which updates the market probability (or possibly better stated, the [normalized price, between 0 and 1](https://forum.effectivealtruism.org/posts/cJc3f4HmFqCZsgGJe/don-t-interpret-prediction-market-prices-as-probabilities)), and a fair payout mechanism which incentivizes risk taking and divergent opinions.

* (1) Formula for Updating Market Probability  - Weighted Probability Adjustment Model (WPAM)
* (2) Market Outcome Update Formulae - Divergence-Based Payout Model (DBPM)

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
& R \in \mathbb{R}, 0 < R < 1 : \text{Resolution probability (ranging from 0 to 1)} & \\
& S \in \mathbb{Z}^+ : \text{Total share pool, sum of all bet amounts} & \\
\end{flalign*}
```

```math
\begin{align*}
& S_{\text{YES}} = \left\lfloor S \cdot R \right\rceil & \\
\end{align*}
```

```math
\begin{align*}
& S_{\text{NO}} = \left\lfloor S \cdot (1 - R) \right\rceil & \\
\end{align*}
```

```math
\begin{flalign*}
& \text{Note that the end result pools S' are produced using Banker's Rounding.} & \\
\end{flalign*}
```

* [Banker's Rounding, Wikipedia](https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even)

---

* It's important to note that the inputs, b_i should be integers, while the probabilities, p_i should be floats. If the p_i is not a float, then there needs to be some kind of ultimate normalized resolution, e.g. between 0 and 100 or 0 and 1000 which represents a probability level p_i or R. So given this, the above calculation will result in fractional shares of S.
* There should ideally be a convention showing what to do with these fractional shares, optionally distribute and always favor one side, randomly favor a side, or create a residual pool which some how gets distributed to users in some manner or acts as a, "tax" on users, perhaps preventing inflation.

##### Step Two - Individual Payout Calculation

* Understanding that we have a series of bets with individual amounts in both YES and NO direction, but we want to incentivize bets that are further away from the R to reward predictors:
* We can introduce d_i, the Reward Factor, representing the linear deviation of the bet's probability p from the resolution probability, R.
* This may be multiplied by the bet amount b_i for any given bet to give us an individual course payout C.

##### DBPM Formula for Calculating Reward Factor

---

```math
\begin{flalign*}
& \text{For Each Bet } i: & \\
\end{flalign*}
```

```math
\begin{align*}
& d_i = |R - p_i| \quad  & \\
& C_i = d_i \times b_i & \\
\end{align*}
```

```math
\begin{flalign*}
& \text{where:} & \\
& R : \text{Resolution probability (ranging from 0 to 1)} & \\
& b_i : \text{Bet amount of bet } i  & \\
& p_i : \text{Market probability at the time of bet } i  & \\
& d_i : \text{the Reward Factor, representing the linear deviation from the R to p at b} & \\
& C_i : \text{he Course payout prior to normalization} & \\
\end{flalign*}
```

---

##### Step Three - Scaling Course Payout to Actual Amount of Capital Pool Available

* The above Step Two may have introduced a problem, where calculating every individual share of every individual better may have either undershot or overshot the amount of capital available in the captial pool (e.g. the actual cash or points available that everyone has thrown into the auction).


```math
\begin{flalign*}
& \text{Step 1: Calculate Course Payouts} & \\
\end{flalign*}
```

```math
\begin{align*}
& C_i \text{ (Course Payout) } = d_i \times b_i & \\
& \text{where } d_i = |R - p_i| & \\
\end{align*}
```

```math
\begin{flalign*}
& \text{Step 2: Sum Course Payouts, C for Each Pool} & \\
\end{flalign*}
```

```math
\begin{align*}
& C_{\text{YES}} = \sum_{i \in \text{YES}} C_i & \\
& C_{\text{NO}} = \sum_{i \in \text{NO}} C_i & \\
\end{align*}
```

```math
\begin{flalign*}
& \text{Step 3: Calculate Normalization Factor, F for Each Pool} & \\
\end{flalign*}
```

```math
\begin{align*}
& F{\text{ (Normalization Factor)}_{\text{YES}}} = \min\left(1, \frac{S_{\text{YES}}}{\text{C}_{\text{YES}}}\right) & \\
& F{\text{ (Normalization Factor)}_{\text{NO}}} = \min\left(1, \frac{S_{\text{NO}}}{\text{C}_{\text{NO}}}\right) & \\
\end{align*}
```

* The above Step 3 does the following:

###### When C>S:

* Use the original normalization to scale down payouts, ensuring that total payouts do not exceed S.

###### When C<S:

* Since scaling payouts with a very small divisor could distort the risk-reward balance that participants agreed upon by massively expanding a payout pool unintentionally, any adjustment to increase C towards S should be done with clear rules and transparency about how these adjustments are made and under what conditions. Therefore our arbitrary rule is to only allow the S/C ratio to ever have a minimum of 1, so we don't get into a scenario where a huge payout due to a small divisor is created unintentionally. Unfortunately this means that actual points can get dropped in the final payout, but this is done to prevent undeserved balooning of payouts.


```math
\begin{flalign*}
& \text{Step 4: Apply Normalization to Calculate Final Payouts} & \\
\end{flalign*}
```

```math
\begin{align*}
& \text{Final Payout}_i = \text{C}_{\text{YES or NO}} \times \text{F}_{\text{YES or NO}} & \\
\end{align*}
```

##### Step Four

```math
\begin{flalign*}
& {Total Payout Distribution, D} & \\
\end{flalign*}
```

```math
\begin{align*}
& D_YES = \sum{\text{Payout}_i \text{ for all YES bets to distribute } S_{\text{YES}}} & \\
& D_NO = \sum{\text{Payout}_i \text{ for all NO bets to distribute } S_{\text{NO}}} & \\
\end{align*}
```