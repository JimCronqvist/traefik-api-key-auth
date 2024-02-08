// Package traefik_api_key_auth Protect your Traefik ingressroutes with API key(s).
package traefik_api_key_auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	//"slices"
)

// Config the plugin configuration.
type Config struct {
	HeaderName            string   `json:"headerName,omitempty"`
	BearerToken           bool     `json:"bearerToken,omitempty"`
	Keys                  []string `json:"keys,omitempty"`
	RemoveHeaderOnSuccess bool     `json:"removeHeaderOnSuccess,omitempty"`
}

// Response the response json when no api key match is found.
type Response struct {
	Message string `json:"message"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName:            "",
		Keys:                  make([]string, 0),
		BearerToken:           false,
		RemoveHeaderOnSuccess: true,
	}
}

// APIKeyAuth a traefik_api_key_auth plugin.
type APIKeyAuth struct {
	next                  http.Handler
	headerName            string
	keys                  []string
	bearerToken           bool
	removeHeaderOnSuccess bool
}

// contains checks if a slice contains a given element (Needed until we can use slices.Contains).
func contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// New created a new APIKeyAuth plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Printf("Creating plugin: %s instance: %+v, ctx: %+v\n", name, *config, ctx)

	if len(config.Keys) == 0 {
		return nil, fmt.Errorf("keys cannot be empty. Please specify at least one API Key")
	}

	// If HeaderName is not provided and BearerToken is true, it defaults to "Authorization".
	if config.HeaderName == "" && config.BearerToken {
		config.HeaderName = "Authorization"
	} else if config.HeaderName == "" {
		config.HeaderName = "X-Api-Key"
	}

	return &APIKeyAuth{
		next:                  next,
		headerName:            config.HeaderName,
		keys:                  config.Keys,
		bearerToken:           config.BearerToken,
		removeHeaderOnSuccess: config.RemoveHeaderOnSuccess,
	}, nil

}

// extractBearerToken Get the Bearer token from the header value.
func extractBearerToken(token string) string {
	re := regexp.MustCompile(`Bearer\s+([^$]+)`)
	match := re.FindStringSubmatch(token)
	if match == nil {
		return ""
	}
	return match[1]
}

func (aka *APIKeyAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check api key header for a valid key, such as x-api-key.
	if !aka.bearerToken {
		if contains(aka.keys, req.Header.Get(aka.headerName)) {
			if aka.removeHeaderOnSuccess {
				req.Header.Del(aka.headerName)
			}
			aka.next.ServeHTTP(rw, req)
			return
		}
	}

	// Check api key header for a valid key in the shape of a Bearer token, such as Authorization.
	if aka.bearerToken {
		bearerToken := extractBearerToken(req.Header.Get(aka.headerName))
		if bearerToken != "" && contains(aka.keys, bearerToken) {
			if aka.removeHeaderOnSuccess {
				req.Header.Del(aka.headerName)
			}
			aka.next.ServeHTTP(rw, req)
			return
		}
	}

	var response Response
	errorMessage := "Invalid API Key. Provide a valid API Key using header %s: %s"
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

	// If no headers or invalid key, return 403.
	if err := json.NewEncoder(rw).Encode(response); err != nil {
		// If response cannot be written, log error.
		fmt.Printf("Error when sending response to an invalid key: %s", err.Error())
	}
}
