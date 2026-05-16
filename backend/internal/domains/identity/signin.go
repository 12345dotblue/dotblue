package identity

import (
	"net/http"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SigninReq is the request body for the signin endpoint
type SigninReq struct {
	Code  string `json:"code" v:"required"`
	State string `json:"state" v:"required"`
}

// SigninRes is the response body for the signin endpoint
type SigninRes struct {
	Token string `json:"token"`
}

// SigninHandler exchanges the OAuth authorization code for a JWT token.
// This endpoint must be called by the frontend on the /callback route.
func SigninHandler(r *ghttp.Request) {
	var req SigninReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	token, err := casdoorsdk.GetOAuthToken(req.Code, req.State)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to exchange OAuth token: %v", err)
		r.Response.WriteStatus(http.StatusUnauthorized, "Failed to exchange token: "+err.Error())
		return
	}

	r.Response.WriteJson(SigninRes{Token: token.AccessToken})
}
