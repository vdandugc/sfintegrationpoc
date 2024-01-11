package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"example.com/sfintegrationpoc/common"
)

type UserInfoResponse struct {
	UserID string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

const (
	loginEndpoint    = "/services/oauth2/token"
	userInfoEndpoint = "/services/oauth2/userinfo"
)

func UserInfo(accessToken string) (*UserInfoResponse, error)  {
	ctx, cancelFn := context.WithTimeout(context.Background(), common.OAuthDialTimeout)
	defer cancelFn()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, common.OAuthEndpoint+userInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code returned on OAuth user info call: %v", httpResp.StatusCode)
	}

	var userInfoResponse UserInfoResponse
	err = json.NewDecoder(httpResp.Body).Decode(&userInfoResponse)

	if err != nil {
		return nil, err
	}
	return &userInfoResponse, nil
}