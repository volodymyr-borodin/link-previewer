package main

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"testing"
)

func TestExtractMeta(t *testing.T) {
	tests := []struct {
		html   string
		result *PageMeta
	}{
		{"<title>Qwerty</title>", &PageMeta{Title: "Qwerty"}},
	}

	for _, tc := range tests {
		t.Run(tc.html, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			result := extractMeta(doc)

			if result.Title != tc.result.Title {
				t.Errorf("got: %s, want: %s", result.Title, tc.result.Title)
			}
		})
	}
}

func TestInputValidationFailed(t *testing.T) {
	tests := []struct {
		m      *InputModel
		result string
	}{
		{nil, "model can't be empty"},
		{&InputModel{Urls: []string{}}, "at least one urls should be specified"},
	}

	for _, tc := range tests {
		t.Run(tc.result, func(t *testing.T) {
			err := tc.m.validate()

			if err.Error() != tc.result {
				t.Errorf("got: %s, want: %s", err.Error(), tc.result)
			}
		})
	}
}

func TestInputValidationSuccess(t *testing.T) {
	tests := []struct {
		m      *InputModel
		result string
	}{
		{&InputModel{Urls: []string{"https://some.com"}}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.result, func(t *testing.T) {
			err := tc.m.validate()

			if err != nil {
				t.Errorf("got: %s, want: %s", err.Error(), "")
			}
		})
	}
}
