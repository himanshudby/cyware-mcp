package ctix

import (
	"github.com/cyware-labs/cyware-mcpserver/common"
	"resty.dev/v3"
)

const Login_endpoint = "rest-auth/login/user-pass/"

type LoginPayload struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	User_id string `json:"user_id"`
	Token   string `json:"token"`
	Email   string `json:"email"`
}

// AuthParams holds the authentication parameters
type AuthParams struct {
	AccessID  string
	Signature string
	Expires   string
}

func generateAuthHeaderForConfig(app common.Application) (string, error) {
	login_resp := LoginResponse{}
	login_payload := LoginPayload{
		Email:    app.Auth.Username,
		Password: app.Auth.Password,
	}
	client := common.APIClient{
		BASE_URL: app.BASE_URL,
		Client:   resty.New(),
	}
	_, err := client.MakeRequest("POST", Login_endpoint, nil, &login_resp, login_payload, nil)
	if err != nil {
		return "", err
	}
	return common.FormatCywareToken(login_resp.Token), nil
}
