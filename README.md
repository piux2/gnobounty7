# gnobounty7

## Problem Definition

https://github.com/gnolang/bounties/issues/28

## Part 1
> TL;DR, I think we want to start with a pure current snapshot of the account state, and apply any changes post-facto.

> part 1: given a block height (the latest blockheight that you know), export the current account state. I believe there is already some feature written to export state (and side note, it probably doesn't use RPC)--used to generate the genesis.json from cosmoshub-3. Test it out at a recent block height. Check it against production data; for example, does it show the amount of tokens held in IBC channels to osmosis? Upload the snapshot to S3 or some other file storage provider.

SOLUTION:

gaiad export --height will export state in json. It can be used as a genesis state to bootstrap a node. 

https://github.com/cosmos/vega-test/blob/66e7ccf559998d48a5bd230a3e6146bed856a83b/exported_unmodified_genesis.json.gz

        curl https://github.com/cosmos/vega-test/blob/66e7ccf559998d48a5bd230a3e6146bed856a83b/exported_unmodified_genesis.json.gz
        unzip ./exported_unmodified_genesis.json.gz

"genesis_time", 2019-12-11T16:11:34Z 

"initial_height", 7,368,387

The state was exported after the Delta (Gravity DEX) 13/07/21 6,910,000 cosmoshub-4 v0.34.x

### Example to find balance on IBC on a 16G RAM, 2.9G CPU computer.

- find all accounts with ibc balances

        time jq '.app_state.bank.balances' exported_unmodified_genesis.json > balances.json

        real	0m19.118s
        user	0m17.170s
        sys	0m1.455s

        time jq  '.[] |select(.coins[].denom != "uatom")' balances.json > ibc_balances.json

        real	0m2.330s
        user	0m2.095s
        sys	0m0.213s

- find accounts holding OSMO tokens  

       time jq  '.[] |select(.coins[].denom == "ibc/14F9BC3E44B8A9C1BE1FB08980FAB87034C9905EF17CF2F5008FC085218811CC")'  -s . ibc_balances.json > ibc_osmo_balances.json

      real	0m0.170s
      user	0m0.144s
      sys	0m0.014s

The token name and denom mapping can be found here [token hub](https://github.com/musicslayer/token_hub/blob/b8f8195ea1f981c11f861f3a26adb37f2dd43500/token_info/atom)


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

Once we dump data in postgreSQL as balances_table and delegation_table. we joined two tables, and add atoms amount and shares amount to get the total atoms that each account owns. 



PROS: a lot more flexible to run SQL against the data once it is imported into the database, especially when we want to retrieve other insights from the same dataset.

CONS: complicated to set up at the beginning. 



For this part of the requirement, it is the same as getting the current balance of each account.
> So if there were no transactions signed by A1 between T1 and T2, (and no unbondings before T1), the result would be simply [{A1;C1}] where C1 are the coins held by A1 at time T1. 

For this part of the requirement, it needs to calculate how many coins are added or removed from the account.  Since the exported state file does not contains individual transactions. We can loop through send, delegation, and unbound messages from the state.db to a postgreSQL database and then query it. 

> given an account A1 at a given block height in the past T1, and current block time T2, create a list of where all the account tokens are now, as a list of {Account;Coins} tuples.

