package main

import (
	"encoding/json"
	"fmt"
	"log"

	"example.com/sfintegrationpoc/common"
	"example.com/sfintegrationpoc/database"
	"example.com/sfintegrationpoc/grpcclient"
	"example.com/sfintegrationpoc/proto"
	"example.com/sfintegrationpoc/rest"
	"github.com/jackc/pgx/v5"
)

var db database.DBConfig = database.DBConfig{
	ConnURL: common.DatabaseConnURL,
}

func main()  {
	if common.ReplayPreset == proto.ReplayPreset_CUSTOM && common.ReplayId == nil {
		log.Fatalf("the replayId variable must be populated when the replayPreset variable is set to CUSTOM")
	} else if common.ReplayPreset != proto.ReplayPreset_CUSTOM && common.ReplayId != nil {
		log.Fatalf("the replayId variable must not be populated when the replayPreset variable is set to EARLIEST or LATEST")
	}

	log.Printf("Creating gRPC client...")
	client, err := grpcclient.NewGRPCClient()
	if err != nil {
		log.Fatalf("could not create gRPC client: %v", err)
	}
	defer client.Close()

	log.Printf("Populating auth token...")
	err = client.Authenticate()
	if err != nil {
		client.Close()
		log.Fatalf("could not authenticate: %v", err)
	}

	log.Printf("Populating user info...")
	err = client.FetchUserInfo()
	if err != nil {
		client.Close()
		log.Fatalf("could not fetch user info: %v",err)
	}

	log.Printf("Making GetTopic request...")
	topic, err := client.GetTopic()
	if err != nil {
		client.Close()
		log.Fatalf("could not fetch topic: %v", err)
	}

	if !topic.GetCanSubscribe() {
		client.Close()
		log.Fatalf("this user is not allowed to subscribe to the following topic: %s", common.TopicName)
	}

	curReplayId := common.ReplayId
	if curReplayId == nil {
		curReplayId, err = db.FetchReplayId()
		if err != nil {
			if err == pgx.ErrNoRows {
				// set replayId to nil in case we weren't able to find it
				curReplayId = nil
			} else {
				log.Fatalf("error in fetching current replay Id: %v", err)
			}
		}
	}

	for {
		log.Printf("Subscribing to topic...")

		// use the user-provided ReplayPreset by default, but if the curReplayId variable has a non-nil value then assume that we want to
		// consume from a custom offset. The curReplayId will have a non-nil value if the user explicitly set the ReplayId or if a previous
		// subscription attempt successfully processed at least one event before crashing
		replayPreset := common.ReplayPreset
		if curReplayId != nil {
			replayPreset = proto.ReplayPreset_CUSTOM
		}

		// channel which will receive events and process them
		eventsChannel := make(chan *proto.ConsumerEvent)
		payloadParserChannel := make(chan map[string]interface{})

		go func() {
			for event := range eventsChannel {
				saveEvent(event)
				payload, err := parseEvent(client, event)
				if err != nil {
					log.Fatalf("error in parsing event: %v", err)
				}
				payloadParserChannel <- payload
			}
		}()

		go func() {
			for payload := range payloadParserChannel {
				fmt.Printf("payload is:  %v", payload)

				var changeEventHeader map[string]interface{} = payload["ChangeEventHeader"].(map[string]interface{})
				entityName := changeEventHeader["entityName"].(string)

				if entityName == "Account" {
					updateAccount(payload)

					// if create then generate signup code
					changeType := changeEventHeader["changeType"].(string)
					if changeType == "CREATE" {
						recordIds := changeEventHeader["recordIds"].([]interface{})

						salesforceAccountId := recordIds[0].(string)
						updateAccountSignupCodeInSalesforce(salesforceAccountId)
					}
				}
			}
		}()

		// In the happy path the Subscribe method should never return, it will just process events indefinitely. In the unhappy path
		// (i.e., an error occurred) the Subscribe method will return both the most recently processed ReplayId as well as the error message.
		// The error message will be logged for the user to see and then we will attempt to re-subscribe with the ReplayId on the next iteration
		// of this for loop
		curReplayId, err = client.Subscribe(replayPreset, curReplayId, eventsChannel)
		if err != nil {
			log.Printf("error occurred while subscribing to topic: %v", err)
		}
	}	
}

func parseEvent(c *grpcclient.PubSubClient, event *proto.ConsumerEvent) (map[string]interface{}, error) {
	codec, err := c.FetchCodec(event.GetEvent().GetSchemaId())
	if err != nil {
		return nil, err
	}

	parsed, _, err := codec.NativeFromBinary(event.GetEvent().GetPayload())
	if err != nil {
		return nil, err
	}

	body, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error casting parsed event: %v", body)
	}

	b, _ := json.Marshal(body)
	fmt.Printf("body: %v", string(b))

	return body, nil
}

// todo: ideally only the `changedFields` in payload should be updated
func updateAccount(data map[string]interface{}) {
	var changeEventHeader map[string]interface{} = data["ChangeEventHeader"].(map[string]interface{})
	recordIds := changeEventHeader["recordIds"].([]interface{})

	id := recordIds[0].(string)
	// can be nil , so extra check for key exists
	var accountNumber *string
	if a, ok := data["AccountNumber"].(*string); ok {
		accountNumber = a
	}

	// extra faff as CREATE event has different structure for name field in payload
	var name *string
	if n, ok := data["Name"].(*string); ok {
		name = n
	} else if nameMap, ok := data["Name"].(map[string]interface{}); ok {
		if n, ok := nameMap["string"].(string); ok {
			name = &n
		}
	}

	/*
	"SignUpCode__c": {
    "string": "TEST123"
  },*/
	var signupCode *string
	if signupCodeMap, ok := data["SignUpCode__c"].(map[string]interface{}); ok {
		if sc, ok := signupCodeMap["string"].(string); ok {
			signupCode = &sc
		}		
	}

	account := database.Account{
		Id: id,
		AccountNumber: accountNumber,
		Name: name,
		SignupCode: signupCode,
	}

	db.UpsertAccount(account)
}

func saveEvent(event *proto.ConsumerEvent) {
	db.SaveEvent(event)
}

func updateAccountSignupCodeInSalesforce(salesforceAccountId string) {
	accountApi := rest.AccountApi{
		AccountId: salesforceAccountId,
	}
	accountApi.UpdateSignupCode()
}