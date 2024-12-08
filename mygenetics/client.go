package mygenetics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/muzykantov/health-gpt/mygenetics/generated"
)

const (
	BaseURL = "https://thsrs-new.sbc.mygenetics.ru"
)

var DefaultClient = &Client{Client: http.DefaultClient}

var (
	ErrUnexpectedType = fmt.Errorf("unexpected type")
	ErrNoFeatures     = fmt.Errorf("no features found")
)

type Client struct {
	*http.Client
}

func (c *Client) Authenticate(
	ctx context.Context,
	email, password string,
) ([]Token, error) {
	const loginURL = BaseURL + "/api/v2/auth/login"

	loginReq := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		loginURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"authentication failed with status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	vals := resp.Header.Values("Set-Cookie")
	if len(vals) == 0 {
		return nil, fmt.Errorf("missing Set-Cookie header")
	}

	out := make([]Token, 0, len(vals))
	for _, val := range vals {
		out = append(out, Token(val))
	}

	return out, nil
}

func (c *Client) Refresh(ctx context.Context, refresh Token) ([]Token, error) {
	const refreshURL = BaseURL + "/api/v2/auth/renew"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, refreshURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Cookie", string(refresh))

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"fetching report failed with status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	vals := resp.Header.Values("Set-Cookie")
	if len(vals) == 0 {
		return nil, fmt.Errorf("missing Set-Cookie header")
	}

	out := make([]Token, 0, len(vals))
	for _, val := range vals {
		out = append(out, Token(val))
	}

	return out, nil
}

func (c *Client) FetchFeatures(
	ctx context.Context,
	access Token,
	codeLab string,
) ([]Feature, error) {
	const reportURL = BaseURL + "/api/v2/codelabs/%s?includeGenes=true&markerFileKey=ru"

	var codelabResponse generated.CodelabResponse
	if err := c.doRequest(
		ctx,
		access,
		http.MethodGet,
		fmt.Sprintf(reportURL, codeLab),
		&codelabResponse,
	); err != nil {
		return nil, err
	}

	var features []Feature
	for featureName := range codelabResponse.Files.Payload.Signs {
		if conclusion, ok := codelabResponse.Files.Payload.Signs[featureName]["conclusion"]; ok {
			conclusionData, ok := conclusion.(map[string]any)
			if !ok {
				continue
			}

			feature, err := c.parseWithConclusion(featureName, conclusionData)
			if err != nil {
				continue
			}

			if len(feature.Genes) == 0 {
				continue
			}

			features = append(features, feature)
		}
	}

	if len(features) > 0 {
		return features, nil
	}

	for _, subFeature := range codelabResponse.Files.Payload.Signs {
		for subFeatureName, subFeature := range subFeature {
			subFeatureData, ok := subFeature.(map[string]any)
			if !ok {
				continue
			}

			feature, err := c.parseWithoutConclusion(
				subFeatureName,
				subFeatureData,
			)
			if err != nil {
				continue
			}

			if len(feature.Genes) == 0 {
				continue
			}

			features = append(features, feature)
		}
	}

	if len(features) > 0 {
		return features, nil
	}

	return nil, ErrNoFeatures
}

func (c *Client) FetchCodelabs(
	ctx context.Context,
	access Token,
) ([]Codelab, error) {
	const reportURL = BaseURL + "/api/v2/tests/"

	var testsResponse generated.TestsResponse
	if err := c.doRequest(
		ctx,
		access,
		http.MethodGet,
		reportURL,
		&testsResponse,
	); err != nil {
		return nil, err
	}

	var codelabs []Codelab
	for _, item := range testsResponse {
		codelabs = append(codelabs, Codelab{
			Code: item.CodeLab,
			Name: item.Profile.Name,
		})
	}

	return codelabs, nil
}

func (c *Client) doRequest(
	ctx context.Context,
	access Token,
	method string,
	url string,
	data any,
) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Cookie", string(access))

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"fetching report failed with status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var rep struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &rep); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if rep.Code != "success" {
		return fmt.Errorf("failed to fetch report: %s", rep.Code)
	}

	// os.WriteFile(fmt.Sprintf("%d.json", time.Now().Unix()), body, 0644)

	if err := json.Unmarshal(rep.Data, data); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}
