package spotify

import "qusic/util"

type QueryType string

const (
	Album     QueryType = "album"
	Artist    QueryType = "artist"
	Playlist  QueryType = "playlist"
	Track     QueryType = "track"
	Show      QueryType = "show"
	Episode   QueryType = "episode"
	Audiobook QueryType = "audiobook"
)

var QueryAll = []QueryType{Album, Artist, Playlist, Track, Show, Episode, Audiobook}

type countryCode string

const (
	AD countryCode = "AD"
	AE countryCode = "AE"
	AF countryCode = "AF"
	AG countryCode = "AG"
	AI countryCode = "AI"
	AL countryCode = "AL"
	AM countryCode = "AM"
	AO countryCode = "AO"
	AQ countryCode = "AQ"
	AR countryCode = "AR"
	AS countryCode = "AS"
	AT countryCode = "AT"
	AU countryCode = "AU"
	AW countryCode = "AW"
	AX countryCode = "AX"
	AZ countryCode = "AZ"
	BA countryCode = "BA"
	BB countryCode = "BB"
	BD countryCode = "BD"
	BE countryCode = "BE"
	BF countryCode = "BF"
	BG countryCode = "BG"
	BH countryCode = "BH"
	BI countryCode = "BI"
	BJ countryCode = "BJ"
	BL countryCode = "BL"
	BM countryCode = "BM"
	BN countryCode = "BN"
	BO countryCode = "BO"
	BQ countryCode = "BQ"
	BR countryCode = "BR"
	BS countryCode = "BS"
	BT countryCode = "BT"
	BV countryCode = "BV"
	BW countryCode = "BW"
	BY countryCode = "BY"
	BZ countryCode = "BZ"
	CA countryCode = "CA"
	CC countryCode = "CC"
	CD countryCode = "CD"
	CF countryCode = "CF"
	CG countryCode = "CG"
	CH countryCode = "CH"
	CI countryCode = "CI"
	CK countryCode = "CK"
	CL countryCode = "CL"
	CM countryCode = "CM"
	CN countryCode = "CN"
	CO countryCode = "CO"
	CR countryCode = "CR"
	CU countryCode = "CU"
	CV countryCode = "CV"
	CW countryCode = "CW"
	CX countryCode = "CX"
	CY countryCode = "CY"
	CZ countryCode = "CZ"
	DE countryCode = "DE"
	DJ countryCode = "DJ"
	DK countryCode = "DK"
	DM countryCode = "DM"
	DO countryCode = "DO"
	DZ countryCode = "DZ"
	EC countryCode = "EC"
	EE countryCode = "EE"
	EG countryCode = "EG"
	EH countryCode = "EH"
	ER countryCode = "ER"
	ES countryCode = "ES"
	ET countryCode = "ET"
	FI countryCode = "FI"
	FJ countryCode = "FJ"
	FK countryCode = "FK"
	FM countryCode = "FM"
	FO countryCode = "FO"
	FR countryCode = "FR"
	GA countryCode = "GA"
	GB countryCode = "GB"
	GD countryCode = "GD"
	GE countryCode = "GE"
	GF countryCode = "GF"
	GG countryCode = "GG"
	GH countryCode = "GH"
	GI countryCode = "GI"
	GL countryCode = "GL"
	GM countryCode = "GM"
	GN countryCode = "GN"
	GP countryCode = "GP"
	GQ countryCode = "GQ"
	GR countryCode = "GR"
	GS countryCode = "GS"
	GT countryCode = "GT"
	GU countryCode = "GU"
	GW countryCode = "GW"
	GY countryCode = "GY"
	HK countryCode = "HK"
	HM countryCode = "HM"
	HN countryCode = "HN"
	HR countryCode = "HR"
	HT countryCode = "HT"
	HU countryCode = "HU"
	ID countryCode = "ID"
	IE countryCode = "IE"
	IL countryCode = "IL"
	IM countryCode = "IM"
	IN countryCode = "IN"
	IO countryCode = "IO"
	IQ countryCode = "IQ"
	IR countryCode = "IR"
	IS countryCode = "IS"
	IT countryCode = "IT"
	JE countryCode = "JE"
	JM countryCode = "JM"
	JO countryCode = "JO"
	JP countryCode = "JP"
	KE countryCode = "KE"
	KG countryCode = "KG"
	KH countryCode = "KH"
	KI countryCode = "KI"
	KM countryCode = "KM"
	KN countryCode = "KN"
	KP countryCode = "KP"
	KR countryCode = "KR"
	KW countryCode = "KW"
	KY countryCode = "KY"
	KZ countryCode = "KZ"
	LA countryCode = "LA"
	LB countryCode = "LB"
	LC countryCode = "LC"
	LI countryCode = "LI"
	LK countryCode = "LK"
	LR countryCode = "LR"
	LS countryCode = "LS"
	LT countryCode = "LT"
	LU countryCode = "LU"
	LV countryCode = "LV"
	LY countryCode = "LY"
	MA countryCode = "MA"
	MC countryCode = "MC"
	MD countryCode = "MD"
	ME countryCode = "ME"
	MF countryCode = "MF"
	MG countryCode = "MG"
	MH countryCode = "MH"
	MK countryCode = "MK"
	ML countryCode = "ML"
	MM countryCode = "MM"
	MN countryCode = "MN"
	MO countryCode = "MO"
	MP countryCode = "MP"
	MQ countryCode = "MQ"
	MR countryCode = "MR"
	MS countryCode = "MS"
	MT countryCode = "MT"
	MU countryCode = "MU"
	MV countryCode = "MV"
	MW countryCode = "MW"
	MX countryCode = "MX"
	MY countryCode = "MY"
	MZ countryCode = "MZ"
	NA countryCode = "NA"
	NC countryCode = "NC"
	NE countryCode = "NE"
	NF countryCode = "NF"
	NG countryCode = "NG"
	NI countryCode = "NI"
	NL countryCode = "NL"
	NO countryCode = "NO"
	NP countryCode = "NP"
	NR countryCode = "NR"
	NU countryCode = "NU"
	NZ countryCode = "NZ"
	OM countryCode = "OM"
	PA countryCode = "PA"
	PE countryCode = "PE"
	PF countryCode = "PF"
	PG countryCode = "PG"
	PH countryCode = "PH"
	PK countryCode = "PK"
	PL countryCode = "PL"
	PM countryCode = "PM"
	PN countryCode = "PN"
	PR countryCode = "PR"
	PS countryCode = "PS"
	PT countryCode = "PT"
	PW countryCode = "PW"
	PY countryCode = "PY"
	QA countryCode = "QA"
	RE countryCode = "RE"
	RO countryCode = "RO"
	RS countryCode = "RS"
	RU countryCode = "RU"
	RW countryCode = "RW"
	SA countryCode = "SA"
	SB countryCode = "SB"
	SC countryCode = "SC"
	SD countryCode = "SD"
	SE countryCode = "SE"
	SG countryCode = "SG"
	SH countryCode = "SH"
	SI countryCode = "SI"
	SJ countryCode = "SJ"
	SK countryCode = "SK"
	SL countryCode = "SL"
	SM countryCode = "SM"
	SN countryCode = "SN"
	SO countryCode = "SO"
	SR countryCode = "SR"
	SS countryCode = "SS"
	ST countryCode = "ST"
	SV countryCode = "SV"
	SX countryCode = "SX"
	SY countryCode = "SY"
	SZ countryCode = "SZ"
	TC countryCode = "TC"
	TD countryCode = "TD"
	TF countryCode = "TF"
	TG countryCode = "TG"
	TH countryCode = "TH"
	TJ countryCode = "TJ"
	TK countryCode = "TK"
	TL countryCode = "TL"
	TM countryCode = "TM"
	TN countryCode = "TN"
	TO countryCode = "TO"
	TR countryCode = "TR"
	TT countryCode = "TT"
	TV countryCode = "TV"
	TW countryCode = "TW"
	TZ countryCode = "TZ"
	UA countryCode = "UA"
	UG countryCode = "UG"
	UM countryCode = "UM"
	US countryCode = "US"
	UY countryCode = "UY"
	UZ countryCode = "UZ"
	VA countryCode = "VA"
	VC countryCode = "VC"
	VE countryCode = "VE"
	VG countryCode = "VG"
	VI countryCode = "VI"
	VN countryCode = "VN"
	VU countryCode = "VU"
	WF countryCode = "WF"
	WS countryCode = "WS"
	XK countryCode = "XK"
	YE countryCode = "YE"
	YT countryCode = "YT"
	ZA countryCode = "ZA"
	ZM countryCode = "ZM"
	ZW countryCode = "ZW"
)

type albumType string

const (
	AlbumTypeAlbum       albumType = "album"
	AlbumTypeSingle      albumType = "single"
	AlbumTypeCompilation albumType = "compilation"
)

type releaseDatePrecision string

const (
	Year  releaseDatePrecision = "year"
	Month releaseDatePrecision = "month"
	Day   releaseDatePrecision = "day"
)

type restrictionReason string

const (
	Market   restrictionReason = "market"
	Product  restrictionReason = "product"
	Explicit restrictionReason = "explicit"
)

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type ImageObject struct {
	URL    string `json:"url"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"width,omitempty"`
}

type SimplifiedArtistObject struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
}

type ArtistObject struct {
	SimplifiedArtistObject
	Followers struct {
		Href  string `json:"href,omitempty"`
		Total int    `json:"total"`
	} `json:"followers"`
	Genres     []string      `json:"genres"`
	Images     []ImageObject `json:"images"`
	Popularity int           `json:"popularity"`
}

type SimplifiedAlbumObject struct {
	AlbumType            albumType            `type:"album_type"`
	TotalTracks          int                  `json:"total_tracks"`
	AvailableMarkets     []countryCode        `json:"available_markets"`
	ExternalURLs         ExternalURLs         `json:"external_urls"`
	Href                 string               `json:"href"`
	ID                   string               `json:"id"`
	Images               []ImageObject        `json:"images"`
	Name                 string               `json:"name"`
	ReleaseDate          string               `json:"release_date"`
	ReleaseDatePrecision releaseDatePrecision `json:"release_date_precision"`
	Restrictions         struct {
		Reason restrictionReason `json:"reason"`
	} `json:"restrictions"`
	URI     string                   `json:"uri"`
	Artists []SimplifiedArtistObject `json:"artists"`
}

type TrackObject struct {
	Album            SimplifiedAlbumObject `json:"album"`
	Artists          []ArtistObject        `json:"artists"`
	AvailableMarkets []countryCode         `json:"available_markets"`
	DiscNumber       int                   `json:"disc_number"`
	DurationMS       int64                 `json:"duration_ms"`
	Explicit         bool                  `json:"explicit"`
	ExternalIDs      struct {
		ISRC string `json:"isrc"`
		EAN  string `json:"ean"`
		UPC  string `json:"upc"`
	} `json:"external_ids"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	IsPlayable   bool         `json:"is_playable"`
	LinkedFrom   *TrackObject `json:"linked_from"`
	Restrictions struct {
		Reason restrictionReason `json:"reason"`
	} `json:"restrictions"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url,omitempty"`
	TrackNumber int    `json:"track_number"`
	URI         string `json:"uri"`
	IsLocal     bool   `json:"is_local"`
}

type SimplifiedPlaylistObject struct {
	Collaborative bool          `json:"collaborative"`
	Description   string        `json:"description"`
	ExternalURLs  ExternalURLs  `json:"external_urls"`
	Href          string        `json:"href"`
	ID            string        `json:"id"`
	Images        []ImageObject `json:"images"`
	Name          string        `json:"name"`
	Owner         struct {
		ExternalURLs ExternalURLs `json:"external_urls"`
		Followers    struct {
			Href  string `json:"href,omitempty"`
			Total int    `json:"total"`
		} `json:"followers"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		URI         string `json:"uri"`
		DisplayName string `json:"display_name,omitempty"`
	} `json:"owner"`
	Public     bool   `json:"public"`
	SnapshotID string `json:"snapshot_id"`
	Tracks     struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	}
	URI string `json:"uri"`
}

type SearchResult struct {
	Tracks  SearchObject[TrackObject]           `json:"tracks"`
	Artists SearchObject[ArtistObject]          `json:"artists"`
	Albums  SearchObject[SimplifiedAlbumObject] `json:"albums"`
}

type SearchObject[T any] struct {
	Href     string `json:"href"`
	Limit    int    `json:"limit"`
	Next     string `json:"next,omitempty"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous,omitempty"`
	Total    int    `json:"total"`
	Items    []T    `json:"items"`
}

type Lyrics struct {
	SyncType string `json:"syncType"`
	Lines    []struct {
		StartTimeMS util.StringInt `json:"startTimeMs"`
		Words       string         `json:"words"`
		Syllables   []any/* idk what type it is */ `json:"syllables"`
		EndTimeMS   util.StringInt `json:"endTimeMs"`
	} `json:"lines"`
	Provider            string `json:"provider"`
	ProviderLyricsId    string `json:"providerLyricsId"`
	ProviderDisplayName string `json:"providerDisplayName"`
	SyncLyricsURI       string `json:"syncLyricsUri"`
	IsDenseTypeface     bool   `json:"isDenseTypeface"`
	Alternatives        []any/* again */ `json:"alternatives"`
	Language            string `json:"language"`
	IsRTLLanguage       bool   `json:"isRtlLanguage"`
	FullScreenAction    string `json:"fullScreenAction"`
	ShowUpsell          bool   `json:"showUpsell"`
	CapStatus           string `json:"capStatus"`
	ImpressionRemaining int    `json:"impressionRemaining"`
	Colors              struct {
		Background    int `json:"background"`
		Text          int `json:"text"`
		HighlightText int `json:"highlightText"`
	} `json:"colors"`
	HasVocalRemoval bool `json:"hasVocalRemoval"`
}
