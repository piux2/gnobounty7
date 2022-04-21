
# gnobounty7

## Problem Definition

https://github.com/gnolang/bounties/issues/28

## Part 2

> part 2: given an account A1 at a given block height in the past T1, and current block time T2, create a list of where all the account tokens are now, as a list of {Account;Coins} tuples. So if there were no transactions signed by A1 between T1 and T2, (and no unbondings before T1), the result would be simply [{A1;C1}] where C1 are the coins held by A1 at time T1. Implementation of this feature would start with SendTx, and then become staking aware. I don't know how best to do that off the top of my head. This also probably shouldn't use RPC but instead use go functions to iterate over blocks to avoid RPC overhead. That said, I might be wrong... if the RPC can handle say a month's worth of cosmoshub-4 blocks through localhost RPC in an hour, then it's fine. This might be feasible with unix pipes, or websockets.


**Observation:**
we can get current snapshot balances of all accounts from the exported state.

    jq  '.app_state.bank.balances |select(.[].coins | length>0)' exported_unmodified_genesis.json

OR an account's balances.

       jq  '.app_state.bank.balances[] |select(.address=="cosmos1z4x4dyylwym26gnsjw29hqjhkwrw6vv2p6t58k")' exported_unmodified_genesis.json

The exported state does not contain user transactions during a period of time
Same for the delegation, it has the current delegation state of each account without the delegation records. 

       jq  '.app_state.staking.delegations'  exported_unmodified_genesis.json > delegations.json

There are  300,587 account

There are  171,027 accounts with balances

There are 144,197 records in the delegations


Because delegated tokens are not part of the account balance, we will need to merge these two together to calculate how many tokens each account owns.  

**Proposal:**

We have two options

[ A ] write a program that just merges the two files in on Json that include delegation share as the additional coin in app_state.bank.balances.coins[] 

Since the dataset is not huge, we merge it in memory. We can sort delegations array by delegations.delegator_address and balances array by address first. We loop through both arrays and put the result in a merged array. 

The time complexity is O(N log N) for quicksort, as the address is quite random, and O(N) for merging 
The space complexity is O(log N) for quicksort and O(1) for merging


PROS: simple

CONS: not flexible. we have to modify the code and get additional insights into the data set. 

#### RESULTS: Less than 3 seconds

Joined and Merged 300,587 Accounts and 144,197 Delegations and tallied staking shares for less than 3s. 

    real	0m2.437s
    user	0m2.085s
    sys	0m0.600s

The source code and data are explained here.

[ B ] dump it to two tables in postgreSQL and write the go program to query the database. 
