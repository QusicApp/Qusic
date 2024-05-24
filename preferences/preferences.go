package preferences

import "fyne.io/fyne/v2"

type preferences map[string]any

func (p preferences) StringWithFallback(key string, fallback string) string {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().StringWithFallback(key, fallback)
		p[key] = v
		return v
	}
	b, _ := v.(string)
	return b
}

func (p preferences) SetString(key string, value string) {
	p[key] = value
	fyne.CurrentApp().Preferences().SetString(key, value)
}

func (p preferences) String(key string) string {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().String(key)
		p[key] = v
		return v
	}
	b, _ := v.(string)
	return b
}

func (p preferences) BoolWithFallback(key string, fallback bool) bool {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().BoolWithFallback(key, fallback)
		p[key] = v
		return v
	}
	b, _ := v.(bool)
	return b
}

func (p preferences) Bool(key string) bool {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().Bool(key)
		p[key] = v
		return v
	}
	b, _ := v.(bool)
	return b
}

func (p preferences) SetBool(key string, value bool) {
	p[key] = value
	fyne.CurrentApp().Preferences().SetBool(key, value)
}

var Preferences = make(preferences)
