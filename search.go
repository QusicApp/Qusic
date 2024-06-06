package main

import (
	_ "image/png"

	"fmt"
	"image"
	"net/http"
	"qusic/logger"
	"qusic/preferences"
	"qusic/widgets"

	pl "qusic/player"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tufteddeer/go-circleimage"
)

func searchPageSkeleton() fyne.CanvasObject {
	songsTxt := canvas.NewText("Songs", theme.ForegroundColor())
	songsTxt.TextSize = theme.TextHeadingSize()
	songsTxt.TextStyle.Bold = true

	songList := container.NewVBox(songsTxt)
	for i := 0; i < 10; i++ {
		rect := canvas.NewRectangle(theme.DisabledButtonColor())
		rect.SetMinSize(fyne.NewSize(0, 48))
		rect.CornerRadius = 5
		songList.Add(rect)
		if i != 9 {
			songList.Add(canvas.NewRectangle(theme.DisabledColor()))
		}
	}

	topResultTxt := canvas.NewText("Top Result", theme.ForegroundColor())
	topResultTxt.TextSize = theme.TextHeadingSize()
	topResultTxt.TextStyle.Bold = true

	topResultRect := canvas.NewRectangle(theme.DisabledButtonColor())
	topResultRect.CornerRadius = 5

	h := 96 + theme.TextSubHeadingSize() + theme.CaptionTextSize()
	topResultRect.SetMinSize(fyne.NewSize(0, h))

	topResult := container.NewVBox(
		topResultTxt,
		topResultRect,
	)

	artistsTxt := canvas.NewText("Artists", theme.ForegroundColor())
	artistsTxt.TextSize = theme.TextHeadingSize()
	artistsTxt.TextStyle.Bold = true
	artists := container.NewGridWrap(fyne.NewSize(150, 150))
	for i := 0; i < 5; i++ {
		artists.Add(container.NewPadded(canvas.NewCircle(theme.DisabledButtonColor())))
	}

	return container.NewGridWithRows(2,
		container.NewGridWithColumns(2,
			container.NewPadded(topResult),
			container.NewVScroll(container.NewPadded(songList)),
		),
		container.NewPadded(container.NewVBox(artistsTxt, container.NewHScroll(container.NewPadded(artists)))),
	)
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
		go func() {
			border.Objects[0] = searchPageSkeleton()
			border.Refresh()
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

					image, err := getImg(song.Thumbnails.Min().URL)
					if err != nil {
						continue
					}
					img := canvas.NewImageFromImage(image)
					img.SetMinSize(fyne.NewSize(48, 48))
					if preferences.Preferences.Bool("hardware_acceleration") {
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

				artistsTxt := canvas.NewText("Artists", theme.ForegroundColor())
				artistsTxt.TextSize = theme.TextHeadingSize()
				artistsTxt.TextStyle.Bold = true
				artists := container.NewHBox()

				for _, artist := range results.Artists {
					name := widget.NewRichText(
						&widget.TextSegment{
							Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
							Text:  artist.Name,
						},
						&widget.TextSegment{
							Style: widget.RichTextStyle{
								ColorName: theme.ColorNamePlaceHolder,
							},
							Text: "Artist",
						},
					)
					name.Truncation = fyne.TextTruncateEllipsis
					img, err := getImg(artist.Thumbnails.Max().URL)
					if err != nil {
						continue
					}

					image := canvas.NewImageFromImage(circleImage(img))
					if preferences.Preferences.Bool("hardware_acceleration") {
						image.ScaleMode = canvas.ImageScaleFastest
					}
					image.FillMode = canvas.ImageFillContain
					image.SetMinSize(fyne.NewSize(150, 150))

					b := container.NewBorder(nil, name, nil, nil, image)
					artists.Add(container.NewPadded(b))
				}

				searchContent = container.NewGridWithColumns(2,
					container.NewPadded(topResult),
					container.NewVScroll(container.NewPadded(songList)),
				)
				searchContent = container.NewGridWithRows(2,
					searchContent,
					container.NewPadded(container.NewVBox(artistsTxt, container.NewHScroll(container.NewPadded(artists)))),
				)
				border.Objects[0] = searchContent
			}

			border.Refresh()
		}()
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

func getImg(url string) (image.Image, error) {
	d, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer d.Body.Close()
	img, _, err := image.Decode(d.Body)
	return img, err
}

func circleImage(img image.Image) image.Image {
	return circleimage.CircleImage(img, image.Point{img.Bounds().Dx() / 2, img.Bounds().Dy() / 2}, img.Bounds().Dx()/2)
}
