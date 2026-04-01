package ctix

import (
	"context"
	"log"

	"github.com/cyware-labs/cyware-mcpserver/common"
	"resty.dev/v3"
)

var (
	ctixDefaultClient common.APIClient
	ctixDefaultConfig common.Application
)

func normalizeCTIXConfig(app common.Application) common.Application {
	app.BASE_URL = common.GetDomain(app.BASE_URL) + "/ctixapi/"
	return app
}

func buildCTIXClient(app common.Application) common.APIClient {
	app = normalizeCTIXConfig(app)

	retryHook := func(r *resty.Response, err error) {
		if r == nil || !common.ContainsStatusCode([]int{400, 401}, r.StatusCode()) {
			return
		}
		switch app.Auth.Type {
		case "basic":
			authToken, err := generateAuthHeaderForConfig(app)
			if err != nil {
				log.Printf("CTIX retry auth error: %v", err)
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

	applyCTIXAuth(app, &client)
	return client
}

func applyCTIXAuth(app common.Application, client *common.APIClient) {
	switch app.Auth.Type {
	case "basic":
		authToken, err := generateAuthHeaderForConfig(app)
		if err != nil {
			log.Printf("CTIX login error: %v", err)
			return
		}
		client.Client.SetHeader("Authorization", authToken)
	case "token":
		token := common.FormatCywareToken(app.Auth.Token)
		client.Client.SetHeader("Authorization", token)
	case "openapicreds":
		params := common.GenerateAuthParams(app.Auth.AccessID, app.Auth.SecretKey)
		client.Client.SetQueryParams(params)
	default:
		log.Printf("unsupported ctix auth_type: %s", app.Auth.Type)
	}
}

func CTIXClientFromContext(ctx context.Context) (common.APIClient, common.Application, bool) {
	if sid, ok := common.SessionIDFromContext(ctx); ok {
		if app, ok := common.GetSessionCTIX(sid); ok && app != nil {
			cfg := normalizeCTIXConfig(*app)
			return buildCTIXClient(cfg), cfg, true
		}
	}
	if ctixDefaultConfig.BASE_URL == "" {
		return common.APIClient{}, common.Application{}, false
	}
	return ctixDefaultClient, ctixDefaultConfig, true
}

