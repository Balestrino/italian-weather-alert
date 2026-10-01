package health

import (
	"context"
	"errors"
	"net/http"
)

type Check func(context.Context) error

// HTTP rejects redirects so private headers cannot travel to another service.
func HTTP(endpoint, token string) Check {
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return errors.New("dependency unavailable")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			return errors.New("dependency unavailable")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.New("dependency unavailable")
		}
		return nil
	}
}
