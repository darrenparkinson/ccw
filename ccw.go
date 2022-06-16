package ccw

import (
	"context"
	"embed"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	//go:embed templates/*
	templates embed.FS
)

type ccwToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time // used to keep track of when the token expires, based on ExpiresIn at the time of creation
}

type tokenError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type Client struct {
	//HTTP Client to use for making requests, allowing the user to supply their own if required.
	HTTPClient *http.Client

	// EstimateService represents the CCW Estimate Service
	EstimateService *EstimateService

	// QuoteService represents the CCW Quote Service
	QuoteService *QuoteService

	username     string
	password     string
	clientID     string
	clientSecret string

	token *ccwToken
	mu    sync.Mutex

	lim *rate.Limiter
}

// EstimateService represents the CCW Quote Service
type EstimateService struct {
	BaseURL string
	client  *Client
}

// QuoteService represents the CCW Quote Service
type QuoteService struct {
	BaseURL string
	client  *Client
}

// NewClient is a helper function that returns an new ccw client given the required parameters.
// Optionally you can provide your own http client or use nil to use the default.  This is done to
// ensure you're aware of the decision you're making to not provide your own http client.
// Each service maintains it's own BaseURL which you can change after calling NewClient if you wish
// to use a different URL.
func NewClient(username, password, clientID, secret string, client *http.Client) (*Client, error) {
	if username == "" || password == "" || clientID == "" || secret == "" {
		return nil, errors.New("missing required parameters")
	}
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	rl := rate.NewLimiter(100, 1)
	c := &Client{
		HTTPClient:   client,
		username:     username,
		password:     password,
		clientID:     clientID,
		clientSecret: secret,
		lim:          rl,
	}

	c.EstimateService = &EstimateService{client: c, BaseURL: "https://api.cisco.com/commerce/EST/v2/async"}
	c.QuoteService = &QuoteService{client: c, BaseURL: "https://api.cisco.com/commerce/QUOTING/v1"}

	return c, nil
}

func (c *Client) makeXMLRequest(ctx context.Context, req *http.Request, v interface{}) error {
	err := c.checkLimitAndGetToken(ctx)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	rc := req.WithContext(ctx)
	res, err := c.HTTPClient.Do(rc)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var ccwErr error
		switch res.StatusCode {
		case 400:
			ccwErr = ErrBadRequest
		case 401:
			ccwErr = ErrUnauthorized
		case 403:
			ccwErr = ErrForbidden
		case 500:
			ccwErr = ErrInternalError
		default:
			ccwErr = ErrUnknown
		}
		return ccwErr
	}
	if res.StatusCode == http.StatusCreated {
		return nil
	}
	if err = xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return err
	}
	return nil
}

func (c *Client) checkLimitAndGetToken(ctx context.Context) error {
	if !c.lim.Allow() {
		c.lim.Wait(ctx)
	}
	err := c.getToken()
	if err != nil {
		return err
	}
	return nil
}

// getToken is a helper function to reuse an existing or retrieve a new token
func (c *Client) getToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != nil && c.token.ExpiresAt.After(time.Now().Add(time.Duration(time.Minute*5))) {
		return nil
	}
	t, err := c.generateToken()
	if err != nil {
		log.Println("error retrieving token")
		return err
	}
	c.token = t
	return nil
}

func (c *Client) generateToken() (*ccwToken, error) {
	u := "https://cloudsso.cisco.com/as/token.oauth2"
	method := "POST"
	username := c.username
	password := c.password
	clientID := c.clientID
	clientSecret := c.clientSecret
	payload := fmt.Sprintf("grant_type=password&username=%s&password=%s&client_id=%s&client_secret=%s", username, password, clientID, clientSecret)
	req, err := http.NewRequest(method, u, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		var e tokenError
		err = json.NewDecoder(res.Body).Decode(&e)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(e.Description)
	}
	var t ccwToken
	err = json.NewDecoder(res.Body).Decode(&t)
	if err != nil {
		return nil, err
	}
	t.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	c.token = &t
	return &t, nil
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	if v != "" {
		return &v
	}
	return nil

}

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// IntOrNil is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it, unless it's zero in
// which case it will return nil. Use this for values you don't
// want to appear in output if they're zero.
func IntOrNil(v int64) *int64 {
	if v != 0 {
		return &v
	}
	return nil
}

// FloatOrNil is a helper routine that allocates a new float64 value
// to store v and returns a pointer to it, unless it's zero in
// which case it will return nil. Use this for values you don't
// want to appear in output if they're zero.
func FloatOrNil(v float64) *float64 {
	if v != 0 {
		return &v
	}
	return nil
}
