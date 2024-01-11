package common

import (
	"time"

	"example.com/sfintegrationpoc/proto"
)

var (
	// topic and subscription-related variables
	TopicName           = "/data/AccountChangeEvent"
	ReplayPreset        = proto.ReplayPreset_EARLIEST
	ReplayId     []byte = nil
	Appetite     int32  = 5

	// gRPC server variables
	GRPCEndpoint    = "api.pubsub.salesforce.com:7443"
	GRPCDialTimeout = 5 * time.Second
	GRPCCallTimeout = 5 * time.Second

	// OAuth header variables
	GrantType    = "password"
	ClientId     = "<CLI_ID>"
	ClientSecret = "<CLI_SECRET>"
	Username     = "<UNAME>"
	Password     = "<PWD>"

	// OAuth server variables
	OAuthEndpoint    = "<URL>"
	OAuthDialTimeout = 5 * time.Second

	AccessToken = "<ACCESS_TOKEN>"
	DatabaseConnURL =  "postgres://foouser:foopassword@localhost:5432/testdb"
	SalesforceAccountObjectURL = OAuthEndpoint + "/services/data/v59.0/sobjects/Account"
)
