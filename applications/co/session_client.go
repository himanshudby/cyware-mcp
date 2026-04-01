package co

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cyware-labs/cyware-mcpserver/common"
	"resty.dev/v3"
)

var (
	coDefaultClient common.APIClient
	coDefaultConfig common.Application

	workspaceBySession sync.Map // sessionID -> workspaceCode
)

func normalizeCOConfig(app common.Application) common.Application {
	app.BASE_URL = common.GetDomain(app.BASE_URL)
	return app
}

func buildCOClient(app common.Application) common.APIClient {
	app = normalizeCOConfig(app)

	retryHook := func(r *resty.Response, err error) {
		if r == nil || !common.ContainsStatusCode([]int{400, 401}, r.StatusCode()) {
			return
		}
		switch app.Auth.Type {
		case "basic":
			authToken, err := generateAuthHeaderForConfig(app)
			if err != nil {
				log.Printf("CO retry auth error: %v", err)
				return
			}
			r.Request.SetHeader("Authorization", authToken)
		case "openapicreds":
			newParams := common.GenerateAuthParams(app.Auth.AccessID, app.Auth.SecretKey)
			r.Request.SetQueryParams(newParams)
		}
	}

	c := common.GetRestyClient(retryHook)
	client := common.APIClient{
		BASE_URL: app.BASE_URL,
		Client:   c,
	}
	applyCOAuth(app, &client)
	return client
}

func applyCOAuth(app common.Application, client *common.APIClient) {
	switch app.Auth.Type {
	case "basic":
		token, err := generateAuthHeaderForConfig(app)
		if err != nil {
			log.Printf("CO login error: %v", err)
			return
		}
		client.Client.SetHeader("Authorization", token)
	case "token":
		token := common.FormatCywareToken(app.Auth.Token)
		client.Client.SetHeader("Authorization", token)
	case "openapicreds":
		params := common.GenerateAuthParams(app.Auth.AccessID, app.Auth.SecretKey)
		client.Client.SetQueryParams(params)
	default:
		log.Printf("unsupported co auth_type: %s", app.Auth.Type)
	}
}

func COClientFromContext(ctx context.Context) (common.APIClient, common.Application, bool) {
	if sid, ok := common.SessionIDFromContext(ctx); ok {
		if app, ok := common.GetSessionCO(sid); ok && app != nil {
			cfg := normalizeCOConfig(*app)
			return buildCOClient(cfg), cfg, true
		}
	}
	if coDefaultConfig.BASE_URL == "" {
		return common.APIClient{}, common.Application{}, false
	}
	return coDefaultClient, coDefaultConfig, true
}

func getWorkspace(ctx context.Context, client common.APIClient) (string, error) {
	if sid, ok := common.SessionIDFromContext(ctx); ok {
		if v, ok := workspaceBySession.Load(sid); ok {
			return v.(string), nil
		}
		ud, err := GetLoggedInUserDetails(ctx)
		if err != nil {
			return "", err
		}
		ws := ud.PreferredWorkspace.Code
		if ws == "" {
			return "", fmt.Errorf("could not determine workspace")
		}
		workspaceBySession.Store(sid, ws)
		return ws, nil
	}
	// No session: treat as default.
	ud, err := GetLoggedInUserDetails(ctx)
	if err != nil {
		return "", err
	}
	ws := ud.PreferredWorkspace.Code
	if ws == "" {
		return "", fmt.Errorf("could not determine workspace")
	}
	return ws, nil
}

func SoarEndpoint(ctx context.Context, endpoint string) (string, error) {
	client, _, ok := COClientFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("CO is not configured for this session")
	}
	ws, err := getWorkspace(ctx, client)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/soarapi/%v/%v", ws, endpoint), nil
}

