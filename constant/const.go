package constant

import "net/url"

type u string

func (strurl u) URL() *url.URL {
	u, _ := url.Parse(string(strurl))
	return u
}

const (
	APP_VERSION   = "Beta 0.1.0"
	GITHUB_URL  u = "https://github.com/QusicApp/Qusic"
	DISCORD_URL u = "https://discord.gg/naVkn4NSXx"
)
