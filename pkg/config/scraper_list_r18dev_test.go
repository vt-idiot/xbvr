package config

import (
	"encoding/json"
	"strings"
	"testing"
)

// The suggestions offered in "Create Custom Site (JAV)" have to save as custom Scrapers as they are:
// an ID no preset has, a source, and the card's numbers.
func TestR18devSuggestions(t *testing.T) {
	var list ScraperList
	if err := json.Unmarshal(officalList, &list); err != nil {
		t.Fatal(err)
	}
	if len(list.XbvrScrapers.R18devSuggestions) == 0 {
		t.Fatal("no suggestions in scrapers.json")
	}
	presets := map[string]bool{}
	for _, p := range list.XbvrScrapers.R18devScrapers {
		presets[strings.ToLower(p.ID)] = true
	}
	seen := map[string]bool{}
	for _, s := range list.XbvrScrapers.R18devSuggestions {
		id := strings.ToLower(s.ID)
		switch {
		case id == "" || presets[id] || seen[id]:
			t.Errorf("suggestion %q: ID empty, a preset's or used twice", s.ID)
		case len(s.Sources) == 0 || len(s.Sites) == 0:
			t.Errorf("suggestion %q: needs sources and sites", s.ID)
		case s.Scenes <= 0 || s.State == "" || s.AvatarUrl == "":
			t.Errorf("suggestion %q: card needs scenes, state and an avatar", s.ID)
		case s.Exclusive || s.AvatarText != "":
			t.Errorf("suggestion %q: quality modes are presets only", s.ID)
		}
		seen[id] = true
	}
}
