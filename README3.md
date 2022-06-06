

## Part 3:
### Problem to solve:

Given a proposal, find all accounts that voted for/against/abstain, but also account for overrides by time (changing votes) and by delegation (overriding the validator's vote by a delegator).

### Basic Concept :
Votes from each account are not exported. Proposals only contain tallied results in an exported state file.

We can loop through vote messages from the tendermint node's state.db to a postgreSQL database and then query it.

The mintscan and tendermint indexer database only contains the votes that are counted towards the final voting power, which requires the delegation to a validator that is bonded among the top 150 or 175 list.

If people voted, but the validators were not bonded at the time the result was tallied, the votes are not included on mintscan website.
In other words, it shows valid votes contributed to the results, not all voters' intent.

We need to retrieve all voting messages and look at who voted, and switch the votes.

To find the account that overrides votes by validators will need further definition. The voters might have multiple delegations to different validators. The voting power is split among validators and proportional to the delegations of account.

If account votes, it will override the validator's vote for sure, regardless validator's options. I think for now, who votes matters more than the different voting options between voters and validators


### Advantage and Limitations

I retrieved the data from QuickSyn

https://quicksync.io/networks/cosmos.html

We don't need to run Gaiad or Tendermint instances.

We can verify the data directly stored in the database that is shared and agreed by the network.

Anyone can download a copy and verify.

You can also read and retrieve the data from your live node.

However, I only tested the data from QuickSyn. It uses golevelDB, and the key prefix is not migrated
to the latest version. It might not work if you use other DB implementations in your nodes. You might run into errors if you run the program against your node with migrated the database key in the nodes.



### Prepare
go 1.8.x

go get github.com/cosmos/cosmos-sdk@v0.46.0-beta1

To fix error:reading github.com/gogo/protobuf/go.mod at revision v1.3.3: unknown revision v1.3.3
Add the following line at the end of the go.mod

    replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

postgres 12.x

	https://www.postgresql.org/download/

Go to the bottom if you setup the postgreSQL the first time.


### Implemenation

In the beginning, I thought we could use temdermint indexer to dump everything and retrieve data from there. However, it will take an estimated more than two days to dump a full archive node to a Postgres indexer on a four-core CPU, 16G mid-sized aws EC2. Plus, the indexed data is still KV pair structure, 99% of transaction data are irrelevant to our problem to solve.


Therefore, I decided to implement something we can be verified with within 2 hours to dump entire data for each full iteration. Once data is dumped to the postsql, each query should take less than 10 seconds.

Go level DB does not support batch insert execution. I implemented a batch insert that significantly improved the performance.

The QuickSyn's block store does not store the record in the same sequences as the original key sequence.
Therefore we will not get correct the base and top height from the database iterator. We need to write our own function to find it.

A log(n) step algorithm to find top and base height.
Find top and base heights in log(n) steps.

All tx hash are stored in Blockstore.db, and Vote transactions are stored in state.db, we will need to map them

Each proposal only has a submitted time and end time. To find the corresponding height of when the proposal is submitted and tallied, we will need to filter through both end block event and deliver_tx

### Result

It took 4.2 hours to download a 2.4T archive node from QuickSync.

applicaton.db, state.db, blockstore.db contains the height of state from height 5200791 (2019-12-11), which is when cosmos-hub3 launched
Here is no application state data before that.

processed 5,271,515 blocks

real    94m37.238s
user    54m50.994s
sys     3m7.935s

600 rows inserted per second
929 blocks per second scanned.


### Instruction

clone github.com/piux2/gnobounty7
cd gnobounty7
make

#### BEFORE you run the extractor program

	./build/extractor extract

please make sure

1) postgreSQL is setup and configure correctly.
2) gnobounty7/config.toml contains correct configuration.
3)  install the contrib modules for postgresql12 . We will need to run cross table function to covert KV table to the relationship table.

	    sudo yum install postgresql12-contrib

      OR

	    sudo apt-get install postgresql12-contrib

4) execute the sql in gnobounty7/psql/schema.sql in your psql client

#### AFTER  ./build/exactor extract


Finally we can get some interesting insights into the data.

login and connect to your database through psql client

All votes are included, not only the votes during the tally where the vote cast through bonded validators.
The votes the people submitted but were not counted towards the end result are also included in the calculation.
I believe that is more accurate.

// find an account that changes votes in the proposal. Please the proposal 38 to a proposal number that you are interested in.

	select * from last_vote
	where proposal ='38' AND sender in (
	      select sender from last_vote where proposal='38'
	      group by sender having count(*) > 1
	  )
	order by proposal
	；



// List Proposals with start and end height

	select * from crosstab_proposal_start_end;

// All all votes of a proposal, replace the proposal id with the one you are interested in.

	select * from crosstab_proposal where proposal = '38'


// You can be creative to join and link tables and heights

## For people who set up PostgreSQL first time.

 Install Postgres and configure the user role and privileges

 - install and  start the service  
 https://www.how2shout.com/linux/install-postgresql-13-on-aws-ec2-amazon-linux-2/
 https://techviewleo.com/install-postgresql-12-on-amazon-linux/

-  Allocate defual data directory
login as postgres : sudo su - postgres
will put you in the postgres data file's grandparent directory

ubuntu: /etc/postgresql/12/main
centos/redhat:/var/lib/pgsql/12/data


- change authentication to md5 hashed password instead of using ident server.

https://serverfault.com/questions/406606/postgres-error-message-fatal-ident-authentication-failed-for-user

Edit pg_hba.conf, relace indent to md5, so that your application can log in to the PostgreSQL DB with password

host all all 127.0.0.1/32 md5
host replication all 127.0.0.1/32 md5

- change the data directory and make sure your mount a big enough disk volume for your postgresql

	  cp  /var/lib/pgsql/12/data/PG_VERSION YOUR_DATA_DIR

edit postgresql-12.service

Environment=PGDATA=YOUR_DATA_DIR


	sudo chown -Rf postgres:postgres /node/psql
	chmod -R 700 /node/psql

- start postgresql service

	  sudo systemctl enable postgresql-12
	  sudo systemctl start postgresql-12

- connect through psql client:

	  login as postgres: sudo su - postgres
	  start clint: psql

- create psql user role for our application talk to postgresql

https://aws.amazon.com/blogs/database/managing-postgresql-users-and-roles/

	CREATE DATABASE cosmos_hub4;

	REVOKE CREATE ON SCHEMA public FROM PUBLIC;
	REVOKE ALL ON DATABASE cosmos_hub4 FROM PUBLIC;

	CREATE ROLE readwrite;
	ALTER ROLE "readwrite" WITH LOGIN;
	GRANT CONNECT ON DATABASE cosmos_hub4 TO readwrite;
	GRANT USAGE, CREATE ON SCHEMA public TO readwrite;


	GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO readwrite;

	GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO readwrite;

	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO readwrite;


	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO readwrite;

	CREATE USER app WITH PASSWORD 'psink';
	GRANT readwrite TO app;

	GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app;
	GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app;
	GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO app;
