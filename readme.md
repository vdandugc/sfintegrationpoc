## Objective

App will sync Account data with Salesforce via its Change Data Capture(CDC) , On Account creation event it will generate a unique code and update Salesforce via API

This is a POC. Items skimped
- Ordering and Buffering (Transaction Based Replication)
- Authentication 
- Security
- Idiomatic Code
- Deletion of record on SF

## Work done 

Listen to events on Topic - `/data/AccountChangedEvent` which Salesforce publishes change data capture events on PubSub when Account Objects are created/updated/deleted to build the Account data on our side.

## Ideally

Initial thoughts on how to approach (can improve as we move forward). Move a pipeline which subscribes on Pub-Sub, store the parse the data and enrich and save the data.

```
subscribe -> store -> parse (group by transactionKey and order by sequence) -> enrich -> save
```

There are other cases which need to be considered and partially considered here.

1. ordering and buffering
2. message reliablity
3. error handling (replayId)

### Ordering and buffering

The order of change events stored in the event bus corresponds to the order in which the transactions corresponding to the record changes are committed in Salesforce. If a transaction includes multiple changes, like a lead conversion, a change event is generated for each change with the same transactionKey but different sequenceNumber in the header. The sequenceNumber is the order of the change within the transaction.

### Message Reliablity and Error Handling

There is a temporary storage in event bus, we also propse to save on our side. Save replayId to replay events in case events are missed, this makes system more resilient. [Link](https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_when_to_use.htm?q=reliability#:~:text=Change%20Data%20Capture%20Reliability)

Refer [this](https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_replication_steps.htm) for error handling.

## Run

To run this
- Set up Salesforce dev account and CLI app (to use the auth token)
- Set up the database schema
- Run 

### Setup Salesforce Account

- Setup account in https://developer.salesforce.com/
- Setup a CLI app ([instructions](https://developer.salesforce.com/docs/atlas.en-us.sfdx_setup.meta/sfdx_setup/sfdx_setup_install_cli.htm#sfdx_setup_install_cli_macos))
- Enable Change Data Capture in Salesforce settings, instructions [here](https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_select_objects.htm)
More on Salesforce Change Data Capture [here](https://developer.salesforce.com/docs/atlas.en-us.246.0.change_data_capture.meta/change_data_capture/cdc_what.htm)
- Update values(`AccessToken` and `OAuthEndpoint`) in common.go

### Schema

Create database with schema

```
>> createdb testdb
>> psql -d testdb
psql >> CREATE USER foouser WITH PASSWORD 'foopassword' CREATEDB CREATEROLE;
psql >> exit

psql -d testdb -U foouser --password

CREATE TABLE ACCOUNTS(
  id TEXT NOT NULL PRIMARY KEY,
  account_number text,
  name text,
  signup_code text
)

CREATE TABLE EVENTS(
  id TEXT NOT NULL PRIMARY KEY,
  schema_id TEXT,
  payload bytea,
  replay_id bytea
)
```

### Run 

```
go run subscribe/subscribe.go
```

