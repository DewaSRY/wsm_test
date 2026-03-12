package marketplace

import (
	"backend/internal/config"
	"backend/internal/domain"
	"bytes"
	"context"
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
	timestamp := time.Now().Unix()
	apiPath := "/oauth/authorize"
	marketPlaceAuth := c.config.MarketPlaceAuth

	params := url.Values{}
	params.Set("shop_id", marketPlaceAuth.ShopId)
	params.Set("state", marketPlaceAuth.State)
	params.Set("partner_id", marketPlaceAuth.PartnerId)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", marketPlaceAuth.Sign)
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

func (c *MarketplaceClient) ExchangeToken(ctx context.Context, code string) (*domain.TokenResponse, error) {
	timestamp := time.Now().Unix()
	apiPath := "/oauth/token"
	marketPlaceAuth := c.config.MarketPlaceAuth

	params := url.Values{}
	params.Set("partner_id", marketPlaceAuth.PartnerId)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", marketPlaceAuth.Sign)

	reqURL := fmt.Sprintf("%s%s?%s", c.config.Marketplace.BaseURL, apiPath, params.Encode())

	body := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
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
	timestamp := time.Now().Unix()
	apiPath := "/oauth/token"
	marketPlaceAuth := c.config.MarketPlaceAuth

	params := url.Values{}
	params.Set("partner_id", marketPlaceAuth.PartnerId)
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", marketPlaceAuth.Sign)

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
		"Authorization": "Bearer " + c.authCode,
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
		"Authorization": "Bearer " + c.authCode,
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
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
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
