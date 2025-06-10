package spotify

// Spotify Web API
// Written by oq 2024
// Some insipiration taken from https://github.com/glomatico/spotify-web-downloader

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/qusicapp/qusic/util"

	widevine "github.com/iyear/gowidevine"
	"github.com/iyear/gowidevine/widevinepb"
)

func New() *Client {
	return new(Client)
}

type Client struct {
	client     http.Client
	id, secret string

	WVDFile      string
	Cookie_sp_dc string

	currentClientID             string
	currentAccessToken          string
	currentAccessTokenExpiry    time.Time
	currentAccessTokenAnonymous bool

	cdm *widevine.CDM
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

func (c *Client) InitWVD() error {
	if c.WVDFile == "" {
		return nil
	}

	f, err := os.Open(c.WVDFile)
	if err != nil {
		return err
	}
	defer f.Close()

	device, err := widevine.NewDevice(widevine.FromWVD(f))
	if err != nil {
		return err
	}
	c.cdm = widevine.NewCDM(device)
	return nil
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
	// Step 1: Fetch server time
	res, err := c.client.Get("https://open.spotify.com/server-time")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch server time")
	}

	var timeResp struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.NewDecoder(res.Body).Decode(&timeResp); err != nil {
		return err
	}

	serverTimeMs := timeResp.ServerTime * 1000
	totp := generateTOTP(serverTimeMs)

	// Step 2: Construct URL with query parameters
	params := url.Values{}
	params.Set("reason", "init")
	params.Set("productType", "web-player")
	params.Set("totp", totp)
	params.Set("totpVer", "5")
	params.Set("ts", strconv.FormatInt(serverTimeMs, 10))

	url := "https://open.spotify.com/get_access_token?" + params.Encode()

	// Step 3: Prepare request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if sp_dc != "" {
		req.Header.Set("Cookie", "sp_dc="+sp_dc)
	}

	res, err = c.client.Do(req)
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

func generateTOTP(timestamp int64) string {
	secret := []byte("5507145853487499592248630329347") // TOTP secret as bytes
	period := int64(30)
	digits := 6

	counter := int64(math.Floor(float64(timestamp) / 1000 / float64(period)))
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], uint64(counter))

	mac := hmac.New(sha1.New, secret)
	mac.Write(counterBytes[:])
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	binCode := (int(hash[offset])&0x7F)<<24 |
		(int(hash[offset+1])&0xFF)<<16 |
		(int(hash[offset+2])&0xFF)<<8 |
		(int(hash[offset+3]) & 0xFF)

	otp := binCode % int(math.Pow10(digits))
	return fmt.Sprintf("%06d", otp)
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
	req.Header.Set("Origin", "https://open.spotify.com")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://open.spotify.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")
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
	req, err := c.newRequest(http.MethodPost, "https://gew4-spclient.spotify.com/widevine-license/v1/audio/license", true, bytes.NewReader(challenge))
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

	if c.cdm == nil {
		return nil, errors.New("WVD is not initialized, check settings")
	}

	challenge, parseLicense, err := c.cdm.GetLicenseChallenge(pssh, widevinepb.LicenseType_AUTOMATIC, false)
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
