package marketplace

import (
	"backend/internal/config"
	"backend/internal/domain"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type MarketplaceClient struct {
	config     *config.Config
	httpClient *http.Client

	// in memory
	authCode     string
	bareToken    string
	refreshToken string
}

func (c *MarketplaceClient) Authorize(ctx context.Context) (*AuthorizeResponse, error) {
	apiPath := "/oauth/authorize"
	marketPlaceAuth := c.config.MarketPlaceAuth

	params, err := c.oauthSignedParams(apiPath, marketPlaceAuth.ShopId)
	if err != nil {
		return nil, err
	}

	params.Set("shop_id", marketPlaceAuth.ShopId)
	params.Set("state", marketPlaceAuth.State)
	params.Set("redirect", marketPlaceAuth.Redirect)

	reqURL := fmt.Sprintf("%s%s?%s", c.config.Marketplace.BaseURL, apiPath, params.Encode())

	var resp AuthorizeResponse
	if err := c.doRequest(ctx, http.MethodGet, reqURL, nil, nil, &resp); err != nil {
		return nil, err
	}

	// store code on in memory
	c.authCode = resp.Data.Code
	return &resp, nil
}

func (c *MarketplaceClient) ExchangeToken(ctx context.Context) (*domain.TokenResponse, error) {
	if c.authCode == "" {
		return nil, fmt.Errorf("authorize code is empty: call Authorize before ExchangeToken")
	}

	apiPath := "/oauth/token"

	params, err := c.oauthSignedParams(apiPath, c.authCode)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s%s?%s", c.config.Marketplace.BaseURL, apiPath, params.Encode())

	body := map[string]string{
		"grant_type": "authorization_code",
		"code":       c.authCode,
	}

	var resp TokenAPIResponse
	if err := c.doRequest(ctx, http.MethodPost, reqURL, nil, body, &resp); err != nil {
		return nil, err
	}

	// store code on in memory
	c.bareToken = resp.Data.AccessToken
	c.refreshToken = resp.Data.RefreshToken
	return &resp.Data, nil
}

func (c *MarketplaceClient) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenResponse, error) {
	apiPath := "/oauth/token"

	params, err := c.oauthSignedParams(apiPath, refreshToken)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s%s?%s", c.config.Marketplace.BaseURL, apiPath, params.Encode())

	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	var resp TokenAPIResponse
	if err := c.doRequest(ctx, http.MethodPost, reqURL, nil, body, &resp); err != nil {
		return nil, err
	}

		// store code on in memory
	c.bareToken = resp.Data.AccessToken
	c.refreshToken = resp.Data.RefreshToken
	return &resp.Data, nil
}

func (c *MarketplaceClient) ListOrders(ctx context.Context) ([]domain.MarketplaceOrder, error) {
	reqURL := fmt.Sprintf("%s/order/list", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + c.bareToken,
	}

	var resp domain.OrderListResponse
	if err := c.doRequest(ctx, http.MethodGet, reqURL, headers, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *MarketplaceClient) GetOrderDetail(ctx context.Context, orderSN string) (*domain.MarketplaceOrder, error) {
	params := url.Values{}
	params.Set("order_sn", orderSN)
	reqURL := fmt.Sprintf("%s/order/detail?%s", c.config.Marketplace.BaseURL, params.Encode())

	headers := map[string]string{
		"Authorization": "Bearer " + c.bareToken,
	}

	var resp struct {
		Message string                  `json:"message"`
		Data    domain.MarketplaceOrder `json:"data"`
	}
	if err := c.doRequest(ctx, http.MethodGet, reqURL, headers, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}


func (c *MarketplaceClient) CancelOrder(ctx context.Context, accessToken, orderSN string) error {
	reqURL := fmt.Sprintf("%s/order/cancel", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	body := map[string]string{
		"order_sn": orderSN,
	}

	var resp domain.APIResponse
	return c.doRequest(ctx, http.MethodPost, reqURL, headers, body, &resp)
}

func (c *MarketplaceClient) UpdateStock(ctx context.Context, accessToken, sku string, quantity int) (*domain.Product, error) {
	reqURL := fmt.Sprintf("%s/product/stock/update", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	body := map[string]interface{}{
		"sku":      sku,
		"quantity": quantity,
	}

	var resp domain.ProductResponse
	if err := c.doRequest(ctx, http.MethodPost, reqURL, headers, body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *MarketplaceClient) UpdatePrice(ctx context.Context, accessToken, sku string, price float64) (*domain.Product, error) {
	reqURL := fmt.Sprintf("%s/product/price/update", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	body := map[string]interface{}{
		"sku":   sku,
		"price": price,
	}

	var resp domain.ProductResponse
	if err := c.doRequest(ctx, http.MethodPost, reqURL, headers, body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *MarketplaceClient) GetLogisticChannels(ctx context.Context, accessToken string) ([]domain.LogisticChannel, error) {
	reqURL := fmt.Sprintf("%s/logistic/channels", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	var resp domain.LogisticChannelsResponse
	if err := c.doRequest(ctx, http.MethodGet, reqURL, headers, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *MarketplaceClient) ShipOrder(ctx context.Context, accessToken, orderSN, channelID string) (*domain.ShipOrderResponse, error) {
	reqURL := fmt.Sprintf("%s/logistic/ship", c.config.Marketplace.BaseURL)

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	body := map[string]string{
		"order_sn":   orderSN,
		"channel_id": channelID,
	}

	var resp struct {
		Message string                   `json:"message"`
		Data    domain.ShipOrderResponse `json:"data"`
	}
	if err := c.doRequest(ctx, http.MethodPost, reqURL, headers, body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}



func (c *MarketplaceClient) doRequest(ctx context.Context, method, reqURL string, headers map[string]string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	marketPlaceConfig := c.config.Marketplace
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	var lastErr error
	for attempt := 0; attempt <= marketPlaceConfig.RetryAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * marketPlaceConfig.RetryDelay
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			log.Printf("Retrying request (attempt %d/%d): %s %s", attempt, marketPlaceConfig.RetryAttempts, method, reqURL)
		}

		if body != nil {
			jsonBody, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode == 401 && c.refreshToken != "" {
			c.RefreshToken(ctx, c.refreshToken)
			continue
		}

		if resp.StatusCode >= 400 {
			var errResp APIErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
				return &APIError{
					StatusCode: resp.StatusCode,
					Message:    errResp.Message,
					ErrorMsg:   errResp.Error,
				}
			}
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(respBody),
			}
		}

		if result != nil {
			trimmedBody := bytes.TrimSpace(respBody)
			if len(trimmedBody) > 0 {
				if err := json.Unmarshal(trimmedBody, result); err != nil {
					// Fallback 1: marketplace may return an envelope and caller expects only `data` payload.
					var envelope struct {
						Data json.RawMessage `json:"data"`
					}
					if envErr := json.Unmarshal(trimmedBody, &envelope); envErr == nil && len(bytes.TrimSpace(envelope.Data)) > 0 {
						if dataErr := json.Unmarshal(envelope.Data, result); dataErr == nil {
							return nil
						}
					}

					// Fallback 2: some gateways return JSON as an escaped string.
					var asString string
					if strErr := json.Unmarshal(trimmedBody, &asString); strErr == nil && asString != "" {
						if unquoteErr := json.Unmarshal([]byte(asString), result); unquoteErr == nil {
							return nil
						}
					}

					return fmt.Errorf("failed to unmarshal response for %s %s: %w; raw=%s", method, reqURL, err, string(trimmedBody))
				}
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", marketPlaceConfig.RetryAttempts+1, lastErr)
}

func New(config *config.Config) *MarketplaceClient {
	return &MarketplaceClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Marketplace.Timeout,
		},
	}
}

func (c *MarketplaceClient) oauthSignedParams(apiPath, suffix string) (url.Values, error) {
	marketPlaceAuth := c.config.MarketPlaceAuth

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sign := ""

	if marketPlaceAuth.PartnerKey != "" {
		sign = c.sign(apiPath, timestamp, suffix)
	} else if marketPlaceAuth.Timestamp != "" && marketPlaceAuth.Sign != "" {
		// Legacy fallback for environments that already provide a pre-signed pair.
		timestamp = marketPlaceAuth.Timestamp
		sign = marketPlaceAuth.Sign
	} else {
		return nil, fmt.Errorf("missing marketplace signing credentials: set PARTNER_KEY or both TIMESTAMP and SIGN")
	}

	params := url.Values{}
	params.Set("partner_id", marketPlaceAuth.PartnerId)
	params.Set("timestamp", timestamp)
	params.Set("sign", sign)

	return params, nil
}

func (c *MarketplaceClient) sign(apiPath, timestamp, suffix string) string {
	marketPlaceAuth := c.config.MarketPlaceAuth

	if marketPlaceAuth.PartnerKey == "" {
		return marketPlaceAuth.Sign
	}

	base := marketPlaceAuth.PartnerId + apiPath + timestamp
	if suffix != "" {
		base += suffix
	}
	h := hmac.New(sha256.New, []byte(marketPlaceAuth.PartnerKey))
	_, _ = h.Write([]byte(base))
	return hex.EncodeToString(h.Sum(nil))
}
