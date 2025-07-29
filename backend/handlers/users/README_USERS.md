### Conventions

#### Convention 20250619 241D9EDE-9C76-4CBB-B08F-D397A74642C5

* Any time an amount is deducted or added to a user balance, we should use apply_transaction.
* There should be consistency in terms of how a user's wallet is changed since it is fundamental to the operation of the entire system to maintain the integrity of number of credits created or destroyed.
* This should apply to payouts, refunds, betting, selling shares, creating markets and so on.

#### Convention 20250619 050E2C8D-3D58-49C5-AEA4-E140FC055A1A

* There should be a uniform way to check user balances prior to performing another operation, similar to 241D9EDE-9C76-4CBB-B08F-D397A74642C5.

#### Convention 20250619 88897BB8-6845-46E4-B43B-0F6652951622

* We should always use the PublicUser model to interact with users for any business logic operation to reduce the risk of exposing user info.