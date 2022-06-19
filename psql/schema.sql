/*
  This file defines the database schema for the PostgresQL ("psql")
  The operator must create a database and installthis schema before using
  the database to index events.
 */

-- The gov table records all events
 CREATE TABLE keyword_event (
  rowid      BIGSERIAL PRIMARY KEY,
  chain_id   VARCHAR NOT NULL,
  height     BIGINT NOT NULL,
  tx_id      INT NOT NULL,
  event_id   INT NOT NULL,
  attribute_id   INT NOT NULL,



  -- type.attributeKey
  composite_key VARCHAR NOT NULL,
  --type
  type VARCHAR NOT NULL,
  --attributeKey
  attribute VARCHAR NOT NULL,
  --attributeValue
  value VARCHAR NOT NULL,

  tx_hash VARCHAR NOT NULL,

  keyword VARCHAR NOT NULL,
  -- When this block header was logged into the sink, in UTC.
  created_at timestamp without time zone default (now() at time zone 'utc'),


  CONSTRAINT chain_height_tx_evet_attribute_value_ckey_hash UNIQUE ( chain_id,height, tx_id, event_id, attribute_id, value, composite_key, tx_hash)
);

CREATE TABLE validator (

val_address VARCHAR NOT NULL,

acc_address VARCHAR NOT NULL,

moniker VARCHAR NOT NULL,

tokens VARCHAR NOT NULL,
shares VARCHAR NOT NULL,
ts_ratio VARCHAR NOT NULL,
height NUMERIC NOT NULL,

CONSTRAINT val UNIQUE (val_address)

);

--every time we create a table we need to grant the access to the app user defiend in our program
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO app;


--login and connect to your database through psql client
CREATE EXTENSION IF NOT EXISTS tablefunc;


-- Covert KV to to relationship table.
  Create VIEW crosstab_proposal AS
  SELECT b.height, a.tx_hash, a.action, a.sender,a.proposal, a.option
  FROM(
    SELECT *
    FROM crosstab(
      'select tx_hash, composite_key, value
       from keyword_event
       Order by 1,2',
       $$VALUES('message.action'::text),('message.sender'::text),('proposal_vote.proposal_id'::text),('proposal_vote.option'::text)$$
      ) AS ct ("tx_hash" text, "action" text,"sender" text,"proposal" text, "option" text)

      Order BY proposal,sender
      ) AS a
    JOIN keyword_event b
    ON a.tx_hash =b.tx_hash
    group by b.height, a.tx_hash, a.action, a.sender,a.proposal,a.option
  ;
-- Get Proposal start and end height.

CREATE VIEW proposal_event AS
SELECT rowid, height, composite_key, value FROM keyword_event WHERE attribute='proposal_result' OR composite_key='active_proposal.proposal_id' OR composite_key='inactive_proposal.proposal_id' OR composite_key='submit_proposal.proposal_id' ORDER BY rowid
;

CREATE VIEW crosstab_proposal_start_end AS
SELECT *
FROM crosstab(
  'select height, composite_key, value
   from proposal_event
   Order by 1',
   $$VALUES('submit_proposal.proposal_id'::text),('active_proposal.proposal_id'::text),('inactive_proposal.proposal_id'::text)$$
  ) AS ct ("height" text, "submit" text, "end" text,"drop" text)
;

-- select unique votes that more than one options, if there is a change vote it will have  more than one entries.
CREATE VIEW more_vote AS
SELECT o.proposal, o.sender, o.option from crosstab_proposal o,
(
  select proposal, sender, count(*) from crosstab_proposal group by proposal, sender having count(*)>1
) AS i
WHERE o.sender = i.sender and o.proposal=i.proposal
GROUP by o.proposal, o.sender, o.option
;
