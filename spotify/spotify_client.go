package spotify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func New() *Client {
	return new(Client)
}

type Client struct {
	client                   http.Client
	id, secret               string
	currentAccessToken       string
	currentAccessTokenExpiry time.Time
}

func (c *Client) expired() bool {
	return time.Now().After(c.currentAccessTokenExpiry)
}

func (c *Client) Ok() bool {
	return c.getAccessToken() == nil
}

func (c *Client) getAccessToken() error {
	res, err := http.Get("https://open.spotify.com/get_access_token?reason=transport&productType=web_player")

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid api key")
	}

	var response struct {
		AccessToken                      string `json:"accessToken"`
		AccessTokenExpirationTimestampMS int64  `json:"accessTokenExpirationTimestampMs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return err
	}
	c.currentAccessToken = response.AccessToken
	c.currentAccessTokenExpiry = time.UnixMilli(response.AccessTokenExpirationTimestampMS)

	return nil
}

func (c *Client) newRequest(method, endpoint string, nobase bool) (*http.Request, error) {
	if c.expired() {
		if err := c.getAccessToken(); err != nil {
			return nil, err
		}
	}

	if !nobase {
		endpoint = "https://api.spotify.com/v1" + endpoint
	}

	req, err := http.NewRequest(method, endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+c.currentAccessToken)
	req.Header.Set("App-Platform", "WebPlayer")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.0.0 Safari/537.36")
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
	req, err := c.newRequest("GET", url, false)
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

// WIP: this is useless until I find a proper way to get client tokens
func (c *Client) Lyrics(trackId string) (Lyrics, error) {
	req, err := c.newRequest("GET",
		fmt.Sprintf("https://spclient.wg.spotify.com/color-lyrics/v2/track/%s?format=json&vocalRemoval=false&market=from_token", url.PathEscape(trackId)),
		true,
	)
	if err != nil {
		return Lyrics{}, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return Lyrics{}, err
	}
	var result struct {
		Lyrics Lyrics `json:"lyrics"`
	}
	err = json.NewDecoder(res.Body).Decode(&result)

	return result.Lyrics, err
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
