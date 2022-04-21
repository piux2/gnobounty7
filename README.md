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
[Continue Here](https://github.com/piux2/gnobounty7/blob/main/README2.md)


