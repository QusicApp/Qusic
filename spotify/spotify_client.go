package spotify

// Spotify Web API
// Written by oq 2024
// Some insipiration taken from https://github.com/glomatico/spotify-web-downloader

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qusic/util"
	"time"

	widevine "github.com/iyear/gowidevine"
	"github.com/iyear/gowidevine/widevinepb"
)

func New() *Client {
	return new(Client)
}

type Client struct {
	client     http.Client
	id, secret string

	Cookie_sp_dc string

	currentClientID             string
	currentAccessToken          string
	currentAccessTokenExpiry    time.Time
	currentAccessTokenAnonymous bool
}

func (c *Client) expired() bool {
	return time.Now().After(c.currentAccessTokenExpiry)
}

func (c *Client) Ok(cookie bool) bool {
	sp_dc := c.Cookie_sp_dc
	if !cookie {
		sp_dc = ""
	}
	return c.getAccessToken(sp_dc) == nil
}

func (c *Client) GetClientToken() (GrantedToken, error) {
	if c.expired() {
		err := c.getAccessToken("")
		if err != nil {
			return GrantedToken{}, err
		}
	}
	var b clientTokenRequest
	b.ClientData.ClientId = c.currentClientID
	b.ClientData.ClientVersion = "1.2.39.110.gcf76504d"
	b.ClientData.JSSDKData = make(map[string]any)
	var body, _ = json.Marshal(b)

	req, _ := http.NewRequest(
		http.MethodPost,
		"https://clienttoken.spotify.com/v1/clienttoken",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return GrantedToken{}, err
	}

	var response clientTokenResponse

	err = json.NewDecoder(res.Body).Decode(&response)

	return response.GrantedToken, err
}

func (c *Client) getAccessToken(sp_dc string) error {
	req, err := http.NewRequest(http.MethodGet, "https://open.spotify.com/get_access_token?reason=transport&productType=web_player", nil)
	if err != nil {
		return err
	}
	if sp_dc != "" {
		req.Header.Set("Cookie", "sp_dc="+sp_dc)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid api key")
	}

	var response struct {
		ClientID                         string `json:"clientId"`
		AccessToken                      string `json:"accessToken"`
		AccessTokenExpirationTimestampMS int64  `json:"accessTokenExpirationTimestampMs"`
		Anonymous                        bool   `json:"isAnonymous"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}
	c.currentAccessToken = response.AccessToken
	c.currentClientID = response.ClientID
	c.currentAccessTokenExpiry = time.UnixMilli(response.AccessTokenExpirationTimestampMS)
	c.currentAccessTokenAnonymous = response.Anonymous

	if response.Anonymous && sp_dc != "" {
		return fmt.Errorf("anonymous token returned for authorized request")
	}

	return nil
}

func (c *Client) newRequest(method, endpoint string, nobase bool, body ...io.Reader) (*http.Request, error) {
	if c.expired() {
		if err := c.getAccessToken(""); err != nil {
			return nil, err
		}
	}

	if !nobase {
		endpoint = "https://api.spotify.com/v1" + endpoint
	}

	var b io.Reader
	if len(body) != 0 {
		b = body[0]
	}

	req, err := http.NewRequest(method, endpoint, b)
	req.Header.Set("Authorization", "Bearer "+c.currentAccessToken)
	req.Header.Set("Accept", "application/json")
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
	req, err := c.newRequest(http.MethodGet, url, false)
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

func (c *Client) TrackMetadata(trackId string) (TrackMetadata, error) {
	req, err := c.newRequest(http.MethodGet,
		fmt.Sprintf("https://spclient.wg.spotify.com/metadata/4/track/%s?market=from_token", trackId),
		true,
	)
	if err != nil {
		return TrackMetadata{}, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return TrackMetadata{}, err
	}
	var metadata TrackMetadata
	err = json.NewDecoder(res.Body).Decode(&metadata)
	return metadata, err
}

func (c *Client) GetAudioFileURLs(fileName string) ([]string, error) {
	req, err := c.newRequest(http.MethodGet,
		fmt.Sprintf("https://gew1-spclient.spotify.com/storage-resolve/v2/files/audio/interactive/10/%s?alt=json", fileName),
		true,
	)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	var response struct {
		Result string   `json:"result"`
		CDNURL []string `json:"cdnurl"`
		FileID string   `json:"fileid"`
		TTL    int      `json:"ttl"`
	}
	err = json.NewDecoder(res.Body).Decode(&response)
	return response.CDNURL, err
}

func (c *Client) Seektable(fileName string) (Seektable, error) {
	res, err := http.Get(fmt.Sprintf("https://seektables.scdn.co/seektable/%s.json", fileName))
	if err != nil {
		return Seektable{}, err
	}
	var response Seektable
	err = json.NewDecoder(res.Body).Decode(&response)
	return response, err
}

func (c *Client) WidevineLicense(challenge []byte) ([]byte, error) {
	req, err := c.newRequest(http.MethodPost, "https://gew1-spclient.spotify.com/widevine-license/v1/audio/license", true, bytes.NewReader(challenge))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	d, err := io.ReadAll(res.Body)

	return d, err
}

func (c *Client) TrackIDToGID(trackId string) string {
	d, _ := util.DecodeBase62(trackId)
	return fmt.Sprintf("%x", d)
}

func (c *Client) GetMP4(trackId string) (*bytes.Reader, error) {
	if err := c.getAccessToken(c.Cookie_sp_dc); err != nil {
		return nil, err
	}
	md, err := c.TrackMetadata(c.TrackIDToGID(trackId))
	if err != nil {
		return nil, err
	}
	file := md.File.Format("MP4_128")

	seektables, err := c.Seektable(file)
	if err != nil {
		return nil, err
	}

	d, err := base64.StdEncoding.DecodeString(seektables.PSSH)
	if err != nil {
		return nil, err
	}

	pssh, err := widevine.NewPSSH(d)
	if err != nil {
		return nil, err
	}

	challenge, parseLicense, err := cdm.GetLicenseChallenge(pssh, widevinepb.LicenseType_AUTOMATIC, false)
	if err != nil {
		return nil, err
	}

	l, err := c.WidevineLicense(challenge)
	if err != nil {
		return nil, err
	}

	keys, err := parseLicense(l)
	if err != nil {
		return nil, err
	}

	urls, err := c.GetAudioFileURLs(file)
	if err != nil {
		return nil, err
	}

	data, err := http.Get(urls[0])
	if err != nil {
		return nil, err
	}

	var buf = new(bytes.Buffer)

	err = widevine.DecryptMP4(data.Body, keys[0].Key, buf)

	return bytes.NewReader(buf.Bytes()), err
}

func (c *Client) Lyrics(trackId string) (Lyrics, error) {
	if err := c.getAccessToken(c.Cookie_sp_dc); err != nil {
		return Lyrics{}, err
	}
	req, err := c.newRequest(http.MethodGet,
		fmt.Sprintf("https://spclient.wg.spotify.com/color-lyrics/v2/track/%s?format=json&vocalRemoval=false&market=from_token", url.PathEscape(trackId)),
		true,
	)
	req.Header.Set("App-Platform", "iOS")
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

const HardcodedWVD = "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QrSyZcBqHO0QrdPcU35GBT7S7+IJr2AAXvnjqnb8yrtpPWN2ZW/IWUJN2z4vZ7/HV4aj3OZhkxC1DIMNyvsusUKoQQuf8gwKiEe8cFwbwFSicywlFk9la2IPe8oFShcxAzHLCCn/TIYUAvEL3/4LgaZvqWm80qCPYbgIP5HT8hPYkKWJ4WYknEWK+3InbnkzteFfGrQFCq4CCAESEGnj6Ji7LD+4o7MoHYT4jBQYjtW+kQUijgIwggEKAoIBAQDY9um1ifBRIOmkPtDZTqH+CZUBbb0eK0Cn3NHFf8MFUDzPEz+emK/OTub/hNxCJCao//pP5L8tRNUPFDrrvCBMo7Rn+iUb+mA/2yXiJ6ivqcN9Cu9i5qOU1ygon9SWZRsujFFB8nxVreY5Lzeq0283zn1Cg1stcX4tOHT7utPzFG/ReDFQt0O/GLlzVwB0d1sn3SKMO4XLjhZdncrtF9jljpg7xjMIlnWJUqxDo7TQkTytJmUl0kcM7bndBLerAdJFGaXc6oSY4eNy/IGDluLCQR3KZEQsy/mLeV1ggQ44MFr7XOM+rd+4/314q/deQbjHqjWFuVr8iIaKbq+R63ShAgMBAAEo8CISgAMii2Mw6z+Qs1bvvxGStie9tpcgoO2uAt5Zvv0CDXvrFlwnSbo+qR71Ru2IlZWVSbN5XYSIDwcwBzHjY8rNr3fgsXtSJty425djNQtF5+J2jrAhf3Q2m7EI5aohZGpD2E0cr+dVj9o8x0uJR2NWR8FVoVQSXZpad3M/4QzBLNto/tz+UKyZwa7Sc/eTQc2+ZcDS3ZEO3lGRsH864Kf/cEGvJRBBqcpJXKfG+ItqEW1AAPptjuggzmZEzRq5xTGf6or+bXrKjCpBS9G1SOyvCNF1k5z6lG8KsXhgQxL6ADHMoulxvUIihyPY5MpimdXfUdEQ5HA2EqNiNVNIO4qP007jW51yAeThOry4J22xs8RdkIClOGAauLIl0lLA4flMzW+VfQl5xYxP0E5tuhn0h+844DslU8ZF7U1dU2QprIApffXD9wgAACk26Rggy8e96z8i86/+YYyZQkc9hIdCAERrgEYCEbByzONrdRDs1MrS/ch1moV5pJv63BIKvQHGvLkaFwoMY29tcGFueV9uYW1lEgd1bmtub3duGioKCm1vZGVsX25hbWUSHEFuZHJvaWQgU0RLIGJ1aWx0IGZvciB4ODZfNjQaGwoRYXJjaGl0ZWN0dXJlX25hbWUSBng4Nl82NBodCgtkZXZpY2VfbmFtZRIOZ2VuZXJpY194ODZfNjQaIAoMcHJvZHVjdF9uYW1lEhBzZGtfcGhvbmVfeDg2XzY0GmMKCmJ1aWxkX2luZm8SVUFuZHJvaWQvc2RrX3Bob25lX3g4Nl82NC9nZW5lcmljX3g4Nl82NDo5L1BTUjEuMTgwNzIwLjAxMi80OTIzMjE0OnVzZXJkZWJ1Zy90ZXN0LWtleXMaHgoUd2lkZXZpbmVfY2RtX3ZlcnNpb24SBjE0LjAuMBokCh9vZW1fY3J5cHRvX3NlY3VyaXR5X3BhdGNoX2xldmVsEgEwMg4QASAAKA0wAEAASABQAA=="

var wvd, _ = base64.StdEncoding.DecodeString(HardcodedWVD)

var device, _ = widevine.NewDevice(widevine.FromWVD(bytes.NewReader(wvd)))

var cdm = widevine.NewCDM(device)
