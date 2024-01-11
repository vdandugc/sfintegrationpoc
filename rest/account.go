package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/sfintegrationpoc/common"
)

type AccountApi struct {
	AccountId string
}

func (a AccountApi) UpdateSignupCode() {
	signupCode := fmt.Sprintf("%s-%d", "SanFrancisco", time.Now().Unix())
	payload, err := json.Marshal(map[string]interface{} {
			"SignUpCode__c": signupCode,
	})
	if err != nil {
			log.Fatal(err)
	}

	client := http.Client{}
	url := fmt.Sprintf("%s/%s", common.SalesforceAccountObjectURL, a.AccountId)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

  req.Header.Set("Content-Type", "application/json")

	token := fmt.Sprintf("Bearer %s", common.AccessToken)
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
			log.Fatal(err)
	}
	log.Println(string(body))
}