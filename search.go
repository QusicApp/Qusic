package main

import (
	"fmt"
	"net/http"
	"qusic/logger"
	"qusic/widgets"

	pl "qusic/player"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func play(i int, w fyne.Window) {
	q := player.Queue()
	s := q[i]
	player.GetVideo(s)

	logger.Inff("Playing song %s: ", s.Name)
	err := player.Play(i)
	logger.Println(err)
	if err != nil {
		return
	}
	setPlayedSong(s, w)
}

func playnow(so *pl.Song, w fyne.Window) {
	player.GetVideo(so)

	logger.Inff("Playing song %s: ", so.Name)
	err := player.PlayNow(so)
	logger.Println(err)
	if err != nil {
		return
	}
	setPlayedSong(so, w)
}

func searchPage(w fyne.Window) fyne.CanvasObject {
	searchBar := widget.NewEntry()
	searchBar.SetPlaceHolder("What do you want to play?")
	searchButton := widgets.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		if searchBar.Text == "" {
			return
		}
		searchBar.OnSubmitted(searchBar.Text)
	})

	border := container.NewBorder(container.NewGridWithColumns(3, layout.NewSpacer(), container.NewBorder(nil, nil, nil, searchButton, searchBar)), nil, nil, nil, searchContent)

	searchBar.OnSubmitted = func(s string) {
		logger.Infof("Searching query \"%s\"", s)
		results := player.Search(s)

		if results.TopResult.ID == "" {
			border.Objects[0] = container.NewCenter(
				widget.NewRichTextFromMarkdown("# SOrry no results :("),
			)
		} else {
			songsTxt := canvas.NewText("Songs", theme.ForegroundColor())
			songsTxt.TextSize = theme.TextHeadingSize()
			songsTxt.TextStyle.Bold = true

			songList := container.NewVBox(songsTxt)
			for i, s := range results.Songs {
				song := s
				image := song.Thumbnails.Min()
				d, _ := http.Get(image.URL)
				img := canvas.NewImageFromReader(d.Body, song.Name)
				img.SetMinSize(fyne.NewSize(48, 48))
				if preferences.Bool("hardware_acceleration") {
					img.ScaleMode = canvas.ImageScaleFastest
				}
				res := &widgets.SongResult{
					Name:           song.Name,
					Artist:         artistText(song.Artists),
					Image:          img,
					DurationString: durString(song.Duration),
					OnTapped: func() {
						playnow(&song, w)
					},
				}
				res.OptionsOnTapped = func() {
					widget.NewPopUpMenu(
						fyne.NewMenu("", fyne.NewMenuItem("Add to queue", func() {
							logger.Infof("Added song to queue: %s", song.Name)
							player.AddToQueue(&song)
						})), w.Canvas(),
					).
						ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(res.Options))
				}
				songList.Add(res)
				if i != len(results.Songs)-1 {
					songList.Add(canvas.NewRectangle(theme.DisabledColor()))
				}
			}

			topResultTxt := canvas.NewText("Top Result", theme.ForegroundColor())
			topResultTxt.TextSize = theme.TextHeadingSize()
			topResultTxt.TextStyle.Bold = true

			topResultRect := canvas.NewRectangle(theme.DisabledColor())
			topResultRect.CornerRadius = 5

			image := results.TopResult.Thumbnails.Max()
			d, _ := http.Get(image.URL)
			topResultImg := canvas.NewImageFromReader(d.Body, results.TopResult.Name)
			topResultImg.SetMinSize(fyne.NewSize(96, 96))
			topResultImg.FillMode = canvas.ImageFillContain

			topResultTitle := canvas.NewText(results.TopResult.Name, theme.ForegroundColor())
			topResultTitle.TextSize = theme.TextSubHeadingSize()
			topResultTitle.TextStyle.Bold = true

			topResultSubtitle := canvas.NewText(fmt.Sprintf("Song • %s", artistText(results.TopResult.Artists)), theme.ForegroundColor())

			topResultOptions := widgets.NewThreeDotOptions(nil)
			topResultOptions.OnTapped = func() {
				widget.NewPopUpMenu(
					fyne.NewMenu("", fyne.NewMenuItem("Add to queue", func() {
						logger.Infof("Added song to queue: %s", results.TopResult.Name)
						player.AddToQueue(&results.TopResult)
					})), w.Canvas(),
				).
					ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(topResultOptions))
			}
			topResult := container.NewVBox(
				topResultTxt,
				container.NewStack(
					topResultRect,
					container.NewPadded(container.NewBorder(nil, nil, container.NewPadded(topResultImg), nil, container.NewVBox(
						layout.NewSpacer(),
						container.NewHBox(topResultTitle, layout.NewSpacer(), topResultOptions),
						layout.NewSpacer(),
						topResultSubtitle,
						layout.NewSpacer(),
						container.NewGridWithColumns(3, &widgets.Button{
							Button: widget.Button{
								Text:       "Play",
								Icon:       theme.MediaPlayIcon(),
								Importance: widget.HighImportance,
								OnTapped: func() {
									playnow(&results.TopResult, w)
								},
							},
						}),
						layout.NewSpacer(),
					))),
				))

			searchContent = container.NewGridWithColumns(2, container.NewPadded(topResult), container.NewVScroll(container.NewPadded(songList)))
			border.Objects[0] = searchContent
		}

		border.Refresh()
	}
	return border
}

func artistText(s []pl.Artist) string {
	var str string
	for i, artist := range s {
		str += artist.Name
		if i != len(s)-1 {
			str += ", "
		}
	}
	return str
}
