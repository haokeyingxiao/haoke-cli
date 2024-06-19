package account_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/haokeyingxiao/haoke-cli/logging"
)

var httpUserAgent = "haoke-cli/0.0.0"

func SetUserAgent(userAgent string) {
	httpUserAgent = userAgent
}

type Client struct {
	Token token `json:"token"`
}

func (c *Client) NewAuthenticatedRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	logging.FromContext(ctx).Debugf("%s: %s", method, path)
	r, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("content-type", "application/json")
	r.Header.Set("accept", "application/json")
	r.Header.Set("X-Shopware-Platform-Token", c.Token.ShopUserToken.Token)
	r.Header.Set("user-agent", httpUserAgent)

	return r, nil
}

func (*Client) doRequest(request *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()

		return nil, fmt.Errorf("doRequest: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("doRequest: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(string(data))
	}

	return data, nil
}

func (c *Client) GetUserID() string {
	return c.Token.UserID
}

func (c *Client) isTokenValid() bool {
	loc, err := time.LoadLocation(c.Token.Expire.Timezone)
	if err != nil {
		return false
	}

	expire, err := time.ParseInLocation("2006-01-02 15:04:05.000000", c.Token.Expire.Date, loc)
	if err != nil {
		return false
	}

	// When it will be expire in the next minute. Respond with false
	return expire.UTC().Sub(time.Now().UTC()).Seconds() > 60
}

const CacheFileName = "haoke-api-client-token.json"

func getApiTokenCacheFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", cacheDir, CacheFileName), nil
}

func createApiFromTokenCache(ctx context.Context) (*Client, error) {
	tokenFilePath, err := getApiTokenCacheFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(tokenFilePath); os.IsNotExist(err) {
		return nil, err
	}

	content, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return nil, err
	}

	var client *Client
	err = json.Unmarshal(content, &client)
	if err != nil {
		return nil, err
	}

	logging.FromContext(ctx).Debugf("Using token cache from %s", tokenFilePath)

	if !client.isTokenValid() {
		return nil, fmt.Errorf("token is expired")
	}

	return client, nil
}

func saveApiTokenToTokenCache(client *Client) error {
	tokenFilePath, err := getApiTokenCacheFilePath()
	if err != nil {
		return err
	}

	content, err := json.Marshal(client)
	if err != nil {
		return err
	}

	tokenFileDirectory := filepath.Dir(tokenFilePath)
	if _, err := os.Stat(tokenFileDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(tokenFileDirectory, 0o750)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(tokenFilePath, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func InvalidateTokenCache() error {
	tokenFilePath, err := getApiTokenCacheFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(tokenFilePath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(tokenFilePath)
}
