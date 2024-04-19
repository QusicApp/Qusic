package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func New(id, secret string) *Client {
	return &Client{id: id, secret: secret, client: http.Client{}}
}

type Client struct {
	id, secret               string
	client                   http.Client
	currentAccessToken       string
	currentAccessTokenExpiry time.Time
}

func (c *Client) expired() bool {
	return time.Now().After(c.currentAccessTokenExpiry)
}

func (c *Client) getAccessToken() error {
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token",
		strings.NewReader(fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", c.id, c.secret)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	var response struct {
		AccessToken string        `json:"access_token"`
		ExpiresIn   time.Duration `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}
	expiry := time.Now().Add(time.Second * response.ExpiresIn)
	c.currentAccessToken = response.AccessToken
	c.currentAccessTokenExpiry = expiry
	return nil
}

func (c *Client) newRequest(method, endpoint string) (*http.Request, error) {
	if c.expired() {
		if err := c.getAccessToken(); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, "https://api.spotify.com/v1"+endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+c.currentAccessToken)
	return req, err
}

func (c *Client) Search(query string, typ []QueryType, market countryCode, limit, offset *int, includeExternalAudio bool) (SearchResult, error) {
	url := fmt.Sprintf("/search?q=%s&type=%s", url.QueryEscape(query), stringsCommaSeperate(typ))
	if market != "" {
		url += "&market=" + string(market)
	}
	if limit != nil {
		url += "&limit=" + fmt.Sprint(*limit)
	}
	if offset != nil {
		url += "&offset=" + fmt.Sprint(*offset)
	}
	if includeExternalAudio {
		url += "&include_external=audio"
	}
	req, err := c.newRequest("GET", url)
	if err != nil {
		return SearchResult{}, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return SearchResult{}, err
	}
	var result SearchResult
	err = json.NewDecoder(res.Body).Decode(&result)
	return result, err
}

func stringsCommaSeperate(s []QueryType) string {
	var str string
	for in, i := range s {
		str += string(i)
		if in != len(s)-1 {
			str += ","
		}
	}
	return str
}
