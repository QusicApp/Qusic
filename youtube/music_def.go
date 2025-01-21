package youtube

import (
	"net/url"
	"github.com/qusicapp/qusic/util"
	"strconv"
	"strings"
	"time"
)

type musicRequest struct {
	Context struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
		} `json:"client"`
	} `json:"context"`
	Params string `json:"params"`
	// for search request
	Query string `json:"query,omitempty"`
	// for playback
	VideoID string `json:"videoId,omitempty"`
	// for browsing
	BrowseID string `json:"browseId,omitempty"`
}

func newMusicRequest(r musicRequest) musicRequest {
	r.Context.Client.ClientName = "WEB_REMIX"
	r.Context.Client.ClientVersion = "1.20240417.01.01"

	return r
}

type musicThumbnailRenderer struct {
	Thumbnail struct {
		Thumbnails []Thumbnail `json:"thumbnails"`
	} `json:"thumbnail"`
}

type runs struct {
	Runs []struct {
		Text               string `json:"text"`
		NavigationEndpoint struct {
			BrowseEndpoint struct {
				BrowseID                              string `json:"browseId"`
				BrowseEndpointContextSupportedConfigs struct {
					BrowseEndpointContextMusicConfig struct {
						PageType string `json:"pageType"`
					} `json:"browseEndpointContextMusicConfig"`
				} `json:"browseEndpointContextSupportedConfigs"`
			} `json:"browseEndpoint"`
			WatchEndpoint struct {
				VideoID string `json:"videoId"`
			} `json:"watchEndpoint"`
		} `json:"navigationEndpoint"`
	} `json:"runs"`
}

func (r runs) Author() []Named {
	authors := make([]Named, 0)
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_ARTIST" || run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_USER_CHANNEL" {
			authors = append(authors, Named{Name: run.Text, ID: run.NavigationEndpoint.BrowseEndpoint.BrowseID})
		}
	}
	return authors
}

func (r runs) Album() Named {
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_ALBUM" {
			return Named{Name: run.Text, ID: run.NavigationEndpoint.BrowseEndpoint.BrowseID}
		}
	}
	return Named{}
}

func (r runs) Duration() time.Duration {
	for _, run := range r.Runs {
		if run.NavigationEndpoint.BrowseEndpoint.BrowseID == "" && strings.Index(run.Text, ":") != -1 {
			dur := run.Text

			sep := strings.Split(dur, ":")
			var (
				hours, minutes, seconds int64
			)
			switch len(sep) {
			case 3:
				hours, _ = strconv.ParseInt(sep[0], 10, 64)
				minutes, _ = strconv.ParseInt(sep[1], 10, 64)
				seconds, _ = strconv.ParseInt(sep[2], 10, 64)
			case 2:
				minutes, _ = strconv.ParseInt(sep[0], 10, 64)
				seconds, _ = strconv.ParseInt(sep[1], 10, 64)
			}

			duration := (time.Duration(hours) * time.Hour) + (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second)
			return duration
		}
	}
	return 0
}

type content struct {
	MusicCardShelfRenderer struct {
		Title     runs `json:"title"`
		Thumbnail struct {
			MusicThumbnailRenderer musicThumbnailRenderer `json:"musicThumbnailRenderer"`
		} `json:"thumbnail"`
		Subtitle runs `json:"subtitle"`
		OnTap    struct {
			WatchEndpoint struct {
				VideoID string `json:"videoId"`
			} `json:"watchEndpoint"`
		} `json:"onTap"`
	} `json:"musicCardShelfRenderer"`
	MusicShelfRenderer struct {
		Title    runs `json:"title"`
		Contents []struct {
			MusicResponsiveListItemRenderer struct {
				PlaylistItemData struct {
					VideoID string `json:"videoId"`
				} `json:"playlistItemData"`
				FlexColumns []struct {
					MusicResponsiveListItemFlexColumnRenderer struct {
						Text runs `json:"text"`
					} `json:"musicResponsiveListItemFlexColumnRenderer"`
				} `json:"flexColumns"`
				Thumbnail struct {
					MusicThumbnailRenderer musicThumbnailRenderer `json:"musicThumbnailRenderer"`
				} `json:"thumbnail"`
			} `json:"musicResponsiveListItemRenderer"`
		} `json:"contents"`
	} `json:"musicShelfRenderer"`
}

type contents []content

func (c contents) Title(s string) content {
	for _, co := range c {
		title := co.MusicShelfRenderer.Title.Runs
		if len(title) == 0 {
			title = co.MusicCardShelfRenderer.Title.Runs
		}
		if len(title) == 0 {
			break
		}
		if title[0].Text == s {
			return co
		}
	}
	return content{}
}

type tab struct {
	TabRenderer struct {
		Title   string
		Content struct {
			SectionListRenderer struct {
				Contents contents `json:"contents"`
			} `json:"sectionListRenderer"`
		} `json:"content"`
	} `json:"tabRenderer"`
}

type tabs []tab

func (t tabs) Title(s string) tab {
	for _, tab := range t {
		if tab.TabRenderer.Title == s {
			return tab
		}
	}
	return tab{}
}

type musicSearchResponse struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs tabs `json:"tabs"`
		} `json:"tabbedSearchResultsRenderer"`
	} `json:"contents"`
}

type Range struct {
	Start util.StringInt `json:"start"`
	End   util.StringInt `json:"end"`
}

type Format struct {
	ApproxDurationMS util.StringInt `json:"approxDurationMs"`
	AverageBitrate   int            `json:"averageBitrate"`
	Bitrate          int            `json:"bitrate"`
	ContentLength    util.StringInt `json:"contentLength"`
	FPS              int            `json:"fps"`
	Height           int            `json:"height"`
	IndexRange       Range          `json:"indexRange"`
	InitRange        Range          `json:"initRange"`
	Itag             int            `json:"itag"`
	LastModified     util.StringInt `json:"lastModified"`
	MimeType         string         `json:"mimeType"`
	ProjectionType   string         `json:"projectionType"`
	Quality          string         `json:"quality"`
	QualityLabel     string         `json:"qualityLabel"`
	SignatureCipher  string         `json:"signatureCipher"`
	Width            int            `json:"width"`
}

func (f Format) URL() string {
	q, _ := url.ParseQuery(f.SignatureCipher)
	return q.Get("url") + "&sig=" + url.QueryEscape(q.Get("s"))
}

type FormatList []Format

type StreamingData struct {
	AdaptiveFormats  FormatList     `json:"adaptiveFormats"`
	ExpiresInSeconds util.StringInt `json:"expiresInSeconds"`
	Formats          FormatList     `json:"formats"`
}

type musicPlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	VideoDetails  struct {
		AllowRating       int            `json:"allowRating"`
		Author            string         `json:"author"`
		ChannelID         string         `json:"channelId"`
		IsCrawlable       bool           `json:"isCrawlable"`
		IsLiveContent     bool           `json:"isLiveContent"`
		IsOwnerViewing    bool           `json:"isOwnerViewing"`
		IsPrivate         bool           `json:"isPrivate"`
		IsUnpluggedCorpus bool           `json:"isUnpluggedCorpus"`
		LengthSeconds     util.StringInt `json:"lengthSeconds"`
		MusicVideoType    string         `json:"musicVideoType"`
		Thumbnail         struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"thumbnail"`
		Title     string         `json:"title"`
		VideoID   string         `json:"videoId"`
		ViewCount util.StringInt `json:"viewCount"`
	} `json:"videoDetails"`

	RequestTime time.Time `json:"-"`
}

func (r musicPlayerResponse) Expires() time.Time {
	e := time.Duration(r.StreamingData.ExpiresInSeconds.Int()) * time.Second

	return r.RequestTime.Add(e)
}

type musicNextResponse struct {
	Contents struct {
		SingleColumnMusicWatchNextResultsRenderer struct {
			TabbedRenderer struct {
				WatchNextTabbedResultsRenderer struct {
					Tabs []struct {
						TabRenderer struct {
							Title    string `json:"title"`
							Endpoint struct {
								BrowseEndpoint struct {
									BrowseID string `json:"browseId"`
								} `json:"browseEndpoint"`
							} `json:"endpoint"`
						} `json:"tabRenderer"`
					} `json:"tabs"`
				} `json:"watchNextTabbedResultsRenderer"`
			} `json:"tabbedRenderer"`
		} `json:"singleColumnMusicWatchNextResultsRenderer"`
	} `json:"contents"`
}

type musicWRBrowseLyricsResponse struct {
	Contents struct {
		SectionListRenderer struct {
			Contents []struct {
				MusicDescriptionShelfRenderer struct {
					Description runs `json:"description"`
					Footer      runs `json:"footer"`
				} `json:"musicDescriptionShelfRenderer"`
			} `json:"contents"`
		} `json:"sectionListRenderer"`
	} `json:"contents"`
}

type TimedLyricData struct {
	LyricLine string `json:"lyricLine"`
	CueRange  struct {
		StartTimeMilliseconds util.StringInt `json:"startTimeMilliseconds"`
		EndTimeMilliseconds   util.StringInt `json:"endTimeMilliseconds"`
		Metadata              struct {
			ID util.StringInt `json:"id"`
		} `json:"metadata"`
	} `json:"cueRange"`
}
type musicIOSBrowseLyricsResponse struct {
	Contents struct {
		ElementRenderer struct {
			NewElement struct {
				Type struct {
					ComponentType struct {
						Model struct {
							TimedLyricsModel struct {
								LyricsData struct {
									TimedLyricsData []TimedLyricData `json:"timedLyricsData"`
									SourceMessage   string           `json:"sourceMessage"`
								} `json:"lyricsData"`
							} `json:"timedLyricsModel"`
						} `json:"model"`
					} `json:"componentType"`
				} `json:"type"`
			} `json:"newElement"`
		} `json:"elementRenderer"`
	} `json:"contents"`
}

type Lyrics struct {
	Source string
	Lyrics string
}
