package main

import (
	"fmt"
	"net/http"
	"qusic/logger"
	"qusic/widgets"
	"qusic/youtube"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func searchPage(w fyne.Window) fyne.CanvasObject {
	searchBar := widget.NewEntry()
	searchBar.SetPlaceHolder(" What do you want to play?")
	searchButton := widgets.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		if searchBar.Text == "" {
			return
		}
		searchBar.OnSubmitted(searchBar.Text)
	})

	border := container.NewBorder(container.NewGridWithColumns(3, layout.NewSpacer(), container.NewBorder(nil, nil, nil, searchButton, searchBar)), nil, nil, nil, searchContent)

	searchBar.OnSubmitted = func(s string) {
		logger.Inff("Searching youtube music (query:\"%s\"): ", s)
		songs, e := client.SearchSongs(s)
		videos, e1 := client.SearchVideos(s)
		fmt.Printf("%d songs (%v), %d videos (%v)\n", len(songs), e, len(videos), e1)

		results := [2][]youtube.MusicSearchResult{songs, videos}
		var forms = [2]fyne.CanvasObject{
			layout.NewSpacer(), layout.NewSpacer(),
		}

		for i, res := range results {
			if len(res) == 0 {
				continue
			}
			txt := "Songs"
			if i == 1 {
				txt = "Videos"
			}
			form := container.NewVBox(widget.NewRichTextFromMarkdown("# " + txt))
			songsc := container.NewVBox()

			for _, s := range res {
				song := s
				image := song.Thumbnails[0]
				d, _ := http.Get(image.URL)
				img := canvas.NewImageFromReader(d.Body, song.Title)
				img.SetMinSize(fyne.NewSize(48, 48))

				songsc.Add(&widgets.SongResult{
					Name:           song.Title,
					Artist:         artistText(song.Authors),
					Image:          img,
					DurationString: durString(song.Duration),
					OnTapped: func() {
						go func() {
							so := player.YoutubeMusicSong(song)
							logger.Inff("Playing song %s: ", so.Name)
							err := player.PlayNow(so)
							fmt.Println(err)
							if err != nil {
								return
							}
							setPlayedSong(so, w)
						}()
					},
				})
			}
			form.Add(songsc)
			forms[i] = form
		}

		grid := container.NewGridWithColumns(2, container.NewVScroll(forms[0]), container.NewVScroll(forms[1]))
		searchContent = grid
		border.Objects[0] = searchContent
		border.Refresh()
	}
	return border
}

func artistText(s []youtube.Author) string {
	var str string
	for i, artist := range s {
		str += artist.Name
		if i != len(s)-1 {
			str += ", "
		}
	}
	return str
}
