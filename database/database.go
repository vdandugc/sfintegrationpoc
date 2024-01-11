package database

import (
	"context"
	"fmt"
	"os"

	"example.com/sfintegrationpoc/proto"
	"github.com/jackc/pgx/v5"
)

type DBConfig struct {
	ConnURL string
}

type Account struct {
	Id string
	AccountNumber *string
	Name *string
	SignupCode *string
}

func (d DBConfig) FetchReplayId() ([]byte, error) {
	conn, err := pgx.Connect(context.Background(), d.ConnURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}

	defer conn.Close(context.Background())

	var replayId []byte
	err = conn.QueryRow(context.Background(), "select replay_id from events order by replay_id desc limit 1").Scan(&replayId)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return nil, err
	}

	return replayId, nil
}

func (d DBConfig) SaveEvent(event *proto.ConsumerEvent) {
	conn, err := pgx.Connect(context.Background(), d.ConnURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `INSERT INTO events (id, schema_id, payload, replay_id) values ($1, $2, $3, $4)`, event.Event.Id, event.Event.SchemaId, event.Event.Payload, event.ReplayId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Insert Query failed: %v\n", err)
		os.Exit(1)
	}
}

func (d DBConfig) UpsertAccount(account Account) {
	// id := "2"
	// object := "{\"testKey\": \"testValue\"}"
	fmt.Println(account.Id, account.AccountNumber, account.Name, account.SignupCode)

	conn, err := pgx.Connect(context.Background(), d.ConnURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	// todo: this is wrong and only for demo, it doesn't update value in table if they are null in Salesforce
	_, err = conn.Exec(context.Background(), `INSERT INTO accounts as a(id, account_number, name, signup_code)
	VALUES($1, $2, $3, $4) 
	ON CONFLICT(id) 
	DO UPDATE SET account_number = COALESCE($2, a.account_number), name = COALESCE($3, a.name), signup_code = COALESCE($4, a.signup_code)`, account.Id, account.AccountNumber, account.Name, account.SignupCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Insert Query failed: %v\n", err)
		os.Exit(1)
	}
}