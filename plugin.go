// Package traefik_api_key_auth Protect your Traefik ingressroutes with API key(s).
package traefik_api_key_auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
)

// Config the plugin configuration.
type Config struct {
	HeaderName            string   `json:"headerName,omitempty"`
	BearerToken           bool     `json:"bearerToken,omitempty"`
	Keys                  []string `json:"keys,omitempty"`
	RemoveHeaderOnSuccess bool     `json:"removeHeaderOnSuccess,omitempty"`
}

// Response the response json when no api key match is found
type Response struct {
	Message string `json:"message"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName:            "X-API-KEY",
		Keys:                  make([]string, 0),
		BearerToken:           true,
		RemoveHeaderOnSuccess: true,
	}
}

// ApiKeyAuth a traefik_api_key_auth plugin.
type ApiKeyAuth struct {
	next                  http.Handler
	headerName            string
	keys                  []string
	bearerToken           bool
	removeHeaderOnSuccess bool
}

// New created a new ApiKeyAuth plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Printf("Creating plugin: %s instance: %+v, ctx: %+v\n", name, *config, ctx)

	if len(config.Keys) == 0 {
		return nil, fmt.Errorf("keys cannot be empty. Please specify at least one API Key")
	}

	// If HeaderName is not provided and BearerToken is true, it defaults to "Authorization".
	if config.HeaderName == "" && config.BearerToken {
		config.HeaderName = "Authorization"
	}

	return &ApiKeyAuth{
		next:                  next,
		headerName:            config.HeaderName,
		keys:                  config.Keys,
		bearerToken:           config.BearerToken,
		removeHeaderOnSuccess: config.RemoveHeaderOnSuccess,
	}, nil

}

// extractBearerToken Get the Bearer token from the header value
func extractBearerToken(token string) string {
	re := regexp.MustCompile(`Bearer\s+([^$]+)`)
	match := re.FindStringSubmatch(token)
	if match == nil {
		return ""
	}
	return match[1]
}

func (aka *ApiKeyAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check api key header for a valid key, such as x-api-key
	if !aka.bearerToken {
		if slices.Contains(aka.keys, req.Header.Get(aka.headerName)) {
			if aka.removeHeaderOnSuccess {
				req.Header.Del(aka.headerName)
			}
			aka.next.ServeHTTP(rw, req)
			return
		}
	}

	// Check api key header for a valid key in the shape of a Bearer token, such as Authorization
	if aka.bearerToken {
		bearerToken := extractBearerToken(req.Header.Get(aka.headerName))
		if bearerToken != "" && slices.Contains(aka.keys, bearerToken) {
			if aka.removeHeaderOnSuccess {
				req.Header.Del(aka.headerName)
			}
			aka.next.ServeHTTP(rw, req)
			return
		}
	}

	var response Response
	errorMessage := "Invalid API Key. Provide an API Key header using %s: %s"
	if aka.bearerToken {
		response = Response{
			Message: fmt.Sprintf(errorMessage, aka.headerName, "Bearer <key>"),
		}
	} else {
		response = Response{
			Message: fmt.Sprintf(errorMessage, aka.headerName, "<key>"),
		}
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusForbidden)

	// If no headers or invalid key, return 403
	if err := json.NewEncoder(rw).Encode(response); err != nil {
		// If response cannot be written, log error
		fmt.Printf("Error when sending response to an invalid key: %s", err.Error())
	}
}
