package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/terraconstructs/grid/pkg/sdk"
)

type sessionConfig struct {
	ServerURL    string
	ClientID     string
	ClientSecret string
}

type tokenProvider struct {
	issuer       string
	clientID     string
	clientSecret string

	mu    sync.Mutex
	creds *sdk.Credentials
}

func (p *tokenProvider) token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.creds != nil && time.Until(p.creds.ExpiresAt) > time.Minute {
		return p.creds.AccessToken, nil
	}

	creds, err := sdk.LoginWithServiceAccount(ctx, p.issuer, p.clientID, p.clientSecret)
	if err != nil {
		return "", fmt.Errorf("service account login failed: %w", err)
	}

	p.creds = creds
	return creds.AccessToken, nil
}

type authTransport struct {
	base     http.RoundTripper
	provider *tokenProvider
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	if t.provider == nil {
		return base.RoundTrip(req)
	}

	token, err := t.provider.token(req.Context())
	if err != nil {
		return nil, err
	}

	clone := req.Clone(req.Context())
	clone.Header = clone.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+token)

	return base.RoundTrip(clone)
}

func newGridClient(ctx context.Context, cfg sessionConfig) (*sdk.Client, error) {
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server url is required (flag --server or GRID_API_URL)")
	}

	authCfg, err := sdk.DiscoverAuthConfig(ctx, cfg.ServerURL)
	if err != nil {
		if isAuthDisabledError(err) {
			return sdk.NewClient(cfg.ServerURL), nil
		}
		return nil, fmt.Errorf("failed to discover auth config: %w", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client credentials are required (flags --client-id/--client-secret or GRID_CLIENT_ID/GRID_CLIENT_SECRET)")
	}

	issuer := authCfg.Issuer
	if issuer == "" {
		issuer = cfg.ServerURL
	}

	httpClient := &http.Client{Transport: &authTransport{
		base:     http.DefaultTransport,
		provider: &tokenProvider{issuer: issuer, clientID: cfg.ClientID, clientSecret: cfg.ClientSecret},
	}}

	return sdk.NewClient(cfg.ServerURL, sdk.WithHTTPClient(httpClient)), nil
}

func isAuthDisabledError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "server returned 404") || strings.Contains(msg, "server returned 503")
}

// ServiceAccountKey represents the structure of a service account JSON file
type ServiceAccountKey struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

var (
	serviceAccountPath string
	outputTokenOnly    bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Grid API and output credentials",
	Long: `Authenticate with Grid API using service account credentials and output access token.
This command is useful for integrating with tools like Atlantis that need to authenticate programmatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Determine credentials source
		var clientID, clientSecret string

		if serviceAccountPath != "" {
			// Read from service account JSON file
			data, err := os.ReadFile(serviceAccountPath)
			if err != nil {
				return fmt.Errorf("failed to read service account file: %w", err)
			}

			var sa ServiceAccountKey
			if err := json.Unmarshal(data, &sa); err != nil {
				return fmt.Errorf("failed to parse service account JSON: %w", err)
			}

			clientID = sa.ClientID
			clientSecret = sa.ClientSecret
		} else {
			// Use flags/environment variables
			clientID = opts.clientID
			clientSecret = opts.clientSecret
		}

		// Validate required fields
		if opts.serverURL == "" {
			return fmt.Errorf("server url is required (flag --server or GRID_API_URL)")
		}
		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("client credentials are required (flags --client-id/--client-secret, --service-account file, or GRID_CLIENT_ID/GRID_CLIENT_SECRET)")
		}

		// Discover auth config
		authCfg, err := sdk.DiscoverAuthConfig(ctx, opts.serverURL)
		if err != nil {
			if isAuthDisabledError(err) {
				return fmt.Errorf("authentication is disabled on the Grid API server")
			}
			return fmt.Errorf("failed to discover auth config: %w", err)
		}

		issuer := authCfg.Issuer
		if issuer == "" {
			issuer = opts.serverURL
		}

		// Authenticate and get token
		creds, err := sdk.LoginWithServiceAccount(ctx, issuer, clientID, clientSecret)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		if outputTokenOnly {
			// Output only the token for easy integration with Atlantis
			fmt.Println(creds.AccessToken)
		} else {
			// Output full credentials information
			fmt.Printf("Successfully authenticated with Grid API\n")
			fmt.Printf("Access Token: %s\n", creds.AccessToken)
			fmt.Printf("Expires At: %s\n", creds.ExpiresAt.Format(time.RFC3339))
		}

		return nil
	},
}

func init() {
	authCmd.Flags().StringVar(&serviceAccountPath, "service-account", "", "Path to service account JSON file (with client_id and client_secret)")
	authCmd.Flags().BoolVar(&outputTokenOnly, "token", false, "Output only the access token (useful for Atlantis integration)")
	rootCmd.AddCommand(authCmd)
}
