# Qusic
Free music app written in Go

## Features
- Downloads songs directly from Spotify! (or YouTube Music)
- No ads
- Lyric support: LRCLIB (synced), YouTube Music (synced), Spotify (synced), Genius, Lyrics.ovh, YouTube Subtitles (synced)

## Installation
### [FFMPEG](https://ffmpeg.org/) is required for the app to work!!

### [Releases](https://github.com/oq-x/qusic/releases)

### Build from source
1. Download [Go](https://go.dev)
2. Download the project, and run `go build`

## Screenshots
![screenshot](screenshots/image.png)
![screenshot](screenshots/image-1.png)

## Use of Spotify cookie
Your spotify cookie is saved locally and is only used to decrypt songs and fetch lyrics.

If you don't enter it, songs will downloaded from YouTube.

Qusic does not interfere with your currently played song on Spotify.

## Acknowledgements
[@dweymouth](https://github.com/dweymouth) - made the synced lyrics module and helped making the YouTube opus streamer

[@xypwn](https://github.com/xypwn) - also helped making the YouTube opus streamer

[kkdai/youtube](https://github.com/kkdai/youtube) - YouTube module used for downloading songs