
# gnobounty7

## Problem Definition

https://github.com/gnolang/bounties/issues/28

## [Part1](https://github.com/piux2/gnobounty7/blob/main/README.md)
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



[ B ] dump it to two tables in postgreSQL and write the go program to query the database.

Once we dump data in postgreSQL as balances_table and delegation_table. we joined two tables, and add atoms amount and shares amount to get the total atoms that each account owns.

PROS: a lot more flexible to run SQL against the data once it is imported into the database, especially when we want to retrieve other insights from the same dataset.

CONS: complicated to set up at the beginning.

For this part of the requirement, it is the same as getting the current balance of each account.
> So if there were no transactions signed by A1 between T1 and T2, (and no unbondings before T1), the result would be simply [{A1;C1}] where C1 are the coins held by A1 at time T1.

For this part of the requirement, it needs to calculate how many coins are added or removed from the account.  Since the exported state file does not contains individual transactions. We can loop through send, delegation, and unbound messages from the state.db to a postgreSQL database and then query it.

> given an account A1 at a given block height in the past T1, and current block time T2, create a list of where all the account tokens are now, as a list of {Account;Coins} tuples.

### SOLUTION A

The app_state.bank.balances contains address and coins[]  (not real data on the chain)
The coins here are the tokens in the wallet.It does not include delegations to validators

            [{
            "address": "cosmos1p2aqt5ux9rquacfjm7ch8h7al00000jjdewgzp",
            "coins": []
          },
          {
            "address": "cosmos1p2a8vx7r00ruz2xmdwm0vk0n000000mng6ccla",
            "coins": [
              {
                "amount": "18513869",
                "denom": "uatom"
              }
            ]
          }]

app_state.staking.delegations contains (not real data on the chain)
The shares are the uatoms in each delgation and on wallet address could have multiple delgations.

     [{
        "delegator_address": "cosmos1p2a8vx7rskruz2xmdwm0vk0n000000mng6ccla",
        "shares": "14180000.000000000000000000",
        "validator_address": "cosmosvaloper100juzk0gdmwu00x4phug7m3ymyatxlh9734g4w"
      },
     {
        "delegator_address": "cosmos1p2a8vx7rskruz2xmdwm0vk0n000000mng6ccla",
        "shares": "2345.000000000000000000",
        "validator_address": "cosmosvaloper12w6ty00j004l8zdla3v4x0jt8lt4rcz5gk7zg2"
      }]
 We will need to join the balances and delegations on address and merge two record to one balance record.
 If there are multi delegations shares we will need to add them togather

 Usage:

     jq '.app_state.bank.balances' exported_unmodified_genesis.json > balances.json
     jq '.app_state.staking.delegations'  exported_unmodified_genesis.json > delegations.json

     git clone https://github.com/piux2/gnobounty7
     cd gnobounty7
     make

     ./build/extract merge --b ../balances.json --d ../delegations.json --va ../validators.json > ../merged.json

     NOTE: We will need the validators.json retrieved from part 3 to take care of validator slashing calcuations. 


 #### RESULTS: Less than 3 seconds

 Joined and Merged 300,587 Accounts and 144,197 Delegations and tallied staking shares for less than 3s.

    real	0m2.437s
    user	0m2.085s
    sys	0m0.600s


To find a account's merged balances.

        jq '.[] | select(.address=="cosmos1p2a8vx7r00ruz2xmdwm0vk0n000000mng6ccla")' merged.json



## Part 3
[Continue Here](https://github.com/piux2/gnobounty7/blob/main/README3.md)
