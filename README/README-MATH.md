## Social Predict Math Discussion

* For background on Prediction Markets, please read [the wikipedia article on Prediction Markets](https://en.wikipedia.org/wiki/Prediction_market) and [this substack article analyzing modern Prediction Market history](https://patdel.substack.com/p/insights-from-recent-prediction-markets).

### Market Mechanisms and Decentralized Finance

* A prediction market is fundamentally a trading platform. But what are we trading? We are trading contracts. Let's say you have two friends, Sven and Ole, and Sven wants to sell an apple to Ole for $1, and they write out a contract for that on a knapkin:

```
Sven promises to deliver a nice apple to Ole tomorrow at 10AM and Ole will give Sven $1.
```

* Sounds simple, right? Wrong! What if Sven and Ole's opinion about what a nice apple is differs? What if Sven is exactly one minute late? What if Ole only has $0.99? Will the contract go through?

* Prediction markets are like creating additional contracts (let's call them Super Contracts) on top of that first contract (call them Original Contracts) to the effect of, "Will that original contract go through?" or, "Will that event happen?"

* However, what's even more fun is not just creating a Super Contract to trade with a friend as a one-off, but rather creating a whole massive collection of Super Contracts to trade among a huge number of strangers...a marketplace of Super Contracts, also known as a Prediction Market Platform.

This sounds cool, but there's a problem. Usually in financial markets like the stock market, there are big participants which help things move and flow. If you're on something like Craigslist or Facebook Marketplace and you sell something like a lamp, you can afford to wait a few days or weeks for that lamp to get a buyer...oh well, you can always still use the lamp. However if you're buying and selling financial products, a stock or an option, you want that thing sold NOW, otherwise you might not want to participate in the marketplace, you might go to another marketplace that will allow you to buy or sell that financial product more immediately.

This introduces the concept of Market Makers and Market Takers, who are participants in a market whose function it is to make sure anything that goes up for sale gets bought roughly immediately.

* Market Makers: Many large financial institutions (like banks, investment firms, or brokerage houses) act as market makers. They provide liquidity to the market by maintaining buy and sell orders on a wide range of financial instruments. They profit from the bid-ask spread.
* Market Takers: These institutions also act as market takers when they trade for their portfolios or on behalf of clients, executing trades against the orders provided by other market makers.

However Prediction Markets are not as fundamental or important as stock markets or bond markets. So rather than having Market Makers and Takers, you need some kind of *other* way to make sure these Super Contracts move and don't just sit there. This might be rules within the software (market mechanisms), meaning in the Prediction Market Platform itself, or there might be some kind of bot that does it, or some other idea.

#### Example Case - Predictia

Imagine a tiny zany island nation called Predictia, which is super far out in the Ocean, so far away that it has its own isolated economy. Predictia is a land full of coconuts, but the citizens who founded Predictia are a bit odd and love betting on things, so rather than creating a coconut backed economy, they instead created a betting-based economy. People still go out and pick coconuts, but rather than bringing them back to the main central government store and selling a coconut for $1 IslandBuck, they instead put the coconut on an auction block, where citizens are able to, "bid," on whether the coconut is either GOOD or BAD, meaning they can enter into a Super Contract onto whether a coconut will be deigned GOOD or BAD.

There is a coconut council which independently and blindly reviews coconuts without knowing their GOOD or BAD bid price, and marks them GOOD or BAD. The pool of IslandBucks from the auction is then re-distributed to all of the participants.

We can say that in the interests of keeping everyone fed, the coconuts are then just divided up communally and eaten by everyone equally. The auction is basically a fun way to move money around in this economy of betting enthusiasts. Anyone has the option to draw IslandBucks from the government bank (which is made of bamboo of course) in the middle of Predictia, taking on personal debt, and then they can test their skill by bidding on coconuts. This is basically Predictia's way of introducing a fiat paper currency, which...could hypothetically help encourage the flow of goods and services, but let's just say it's more about having something to do on a Friday night.

##### Consequences of IslandBucks and the Coconut Auction

* Winners and Losers: This economy creates a distinct division between winners (who correctly judge the coconuts and gain shares of the pooled IslandBucks) and losers (who get nothing and potentially lose their borrowed IslandBucks). This could lead to wealth disparities and financial risk for participants.

* Economic Implications:
* Inflation Risk: The issuance of new currency to winners could lead to inflation, especially if the amount of currency in circulation grows faster than the economic activity or productivity, because so many auctions are put on because they are so fun.
* Debt Management: Predictia needs a system to manage and track the debts of its citizens, including mechanisms for repayment and possibly interest.
* Resource Allocation: The communal consumption of the coconut, irrespective of the auction outcome, suggests a communal or shared aspect of some resources, despite the competitive bidding process.

###### Accounting for Debt and New Currency:

There are some solutions to the above problems:

* For the Government: New debts (IslandBucks borrowed by citizens) would be recorded as liabilities, and the corresponding amount of IslandBucks issued would be part of the government's assets.
* For Citizens: Borrowed IslandBucks become a personal liability, and any gains from the auction are assets.
* Predictia introduces a credit system where citizens can borrow IslandBucks up to a certain limit (let's say 500 IslandBucks). This system creates debt obligations for the citizens, who must presumably repay the borrowed currency at some point.
* IslandBucks function as a fiat currency because their value is not inherently backed by a physical commodity. Instead, their value is based on the trust and acceptance of the participants in the economy. So if the auction is not done well, if trust breaks down, then the whole system gets thrown out because people see it as an unfair sham.

###### Rationale for Market Mechanisms in Predictia

###### The Role of Central Authorities in Traditional Finance

In traditional finance, central authorities mentioned above play a significant role:

* Regulatory Bodies: These include entities like the SEC (U.S. Securities and Exchange Commission) or other financial regulatory agencies. They oversee market operations, enforce regulations, and aim to ensure fairness and stability.
* Central Banks: They influence liquidity in the broader economy through monetary policy, interest rates, etc.
* Large Financial Institutions: Banks and other large financial institutions can also act as major market makers.

On Predictia, we don't have a complex society, it's only let's say...100 people. So rather than hainv large financial instutitions which act as Market Makers and Takers, you instead have software and market mechanisms to take care of the coconut Super Contracts to make sure there is a constant flow if someone wants to buy or sell a Super Contract.

This is basically why our platform, SocialPredict has to have some kind of underlying math and market mechanisms, to support the ongoing flow of buying and selling YES and NO contracts!

### Market Probability Update Formulae

#### Formula for Updating Market Probability - Weighted Probability Adjustment Model (WPAM)

* The Weighted Probability Adjustment Model is our base model and the math functions for this can be viewed by searching our codebase for WPAM. WPAM is designed to update the probability of an outcome based on the total amount bet on each possibility. It gives more weight to the initial settings (initial probability and investment) to stabilize the market in its early phase.

WPAM Formula:

```
P_{\text{new}} = \frac{P_{\text{initial}} \times I_{\text{initial}} + A_{\text{YES}}}{I_{\text{initial}} + A_{\text{YES}} + A_{\text{NO}}}

\text{where:} \\
P_{\text{new}} \text{ is the new probability.} \\
P_{\text{initial}} \text{ is the initial probability, set to 0.5.} \\
I_{\text{initial}} \text{ is the initial investment, assumed to be 10 points.} \\
A_{\text{YES}} \text{ is the total amount bet on "YES".} \\
A_{\text{NO}} \text{ is the total amount bet on "NO".}
```

* Initial Probability as a Weighted Factor: The initial probability (P_initial) is typically set to represent a balanced or neutral starting point for the market, often 0.5 for a 50-50 chance. This value is used as a weighted factor in the numerator to establish the baseline influence on the market's direction.
* The cost of creating the market is a way of weighting the market as an initial investment (I_initial). While the (P_initial) is meant to represent blind uncertainty, a 50-50 chance of any market created, the (I_initial) is meant to represent a form of stability, which is why it is included in both the numerator and denominator. If there is a larger initial investment, such that (I_initial) >> (A_YES) or (A_NO)  this implies that the market will not move as much until larger (A_YES) or (A_NO) is invested.

#### Example Orders and Outcomes

1. (I_initial) of 10, (P_initial) of 0.50, (A_NO) order made in amount of 20.

```
P_{\text{new}} = \frac{0.5 \times 10}{10 + 0 + 20} = \frac{5}{30} \approx 0.167
```

2. Same as above but (A_YES) order made in amount of 10.

```
P_{\text{new}} = \frac{0.5 \times 10 + 10}{10 + 10 + 0} = \frac{15}{20} = 0.75
```

3. Follow up order on (2) made in (A_NO) direction.

```
P_{\text{new}} = \frac{0.5 \times 10}{10 + 0 + 20} = \frac{5}{30} \approx 0.167
```

#### Formula for Updating Market Probability - Constant Product Market Maker (CPMM)

* An alternative to WPAM is CPMM, which is a form of [Function Product Market Making](https://en.wikipedia.org/wiki/Constant_function_market_maker), meaning there is some kind of underlying function that has to maintain state. This is a method that was used in cryptocurrencies, who do not have market makers.

One example CPMM is the [constant product rule used by Uniswap](https://en.wikipedia.org/wiki/Uniswap).

Bots or individuals, termed "liquidity providers"—provide liquidity to the exchange by adding a pair of tokens to a smart contract which can be bought and sold by other users following the rule:

```
k = xy
```

* Where k is a constant ratio of the price of two cryptocurrencies.
* Liquidity providers are given a percentage of the trading fees earned for that trading pair, so there is a built-in incentive to provide liquidity.

While CPMM isn't directly used for updating probabilities in a binary market, it can indirectly influence them:

* Market Dynamics: The way CPMM affects prices can impact market participants' perceptions and their betting behavior. As participants react to price changes and adjust their bets, this could indirectly influence the calculated probabilities of outcomes in a prediction market.
* Indirect Influence: If a prediction market uses a CPMM model, the changes in the asset quantities (representing “YES” and “NO” positions, for example) could be used to infer changes in market sentiment, which might then be used to update probabilities. However, this would be an indirect effect and would require additional modeling to translate price or ratio changes into probabilities.

So basically the while the introduction of CPMM into a marketplace wouldn't directly define the formula showing how a probability is updated, it could affect user behavior because users know that there is a bot buying and selling shares to ensure that an underlying consistent state is maintained.

### Market Outcome Update Formulae

#### Weighted Probability Adjustment Model (WPAM)