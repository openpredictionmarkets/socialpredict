# Economics Configuration

* The economics of SocialPredict can be customized upon set up based upon entries in a YAML file.
* Elements such as the cost to create a market, what the initial probability of a market is set to, how much a market is subsidized by the computer, trader bonuses and so on can be set up within this file.
* These are global variables which effect the operation of the entire software suite and for now are meant to be set up permanently for the entirety of the run of a particular instance of SocialPredict.

```
economics:
  marketcreation:
    initialMarketProbability: 0.5
    initialMarketSubsidization: 10
    initialMarketYes: 0
    initialMarketNo: 0
  marketincentives:
    createMarketCost: 1
    traderBonus: 2
  user:
    initialAccountBalance: 0
    maximumDebtAllowed: 500
  betting:
    minimumBet: 1
    betFee: 0
    sellSharesFee: 0
```

* We may implement variable economics in the future, however this might need to come along with transparency metrics, which show how the economics were changed to users, which requires another level of data table to be added.