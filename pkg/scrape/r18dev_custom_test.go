package scrape

import (
	"testing"

	"github.com/xbapps/xbvr/pkg/config"
)

func TestR18devSourceFromURL(t *testing.T) {
	ok := map[string]config.R18devSource{
		"https://r18.dev/videos/vod/movies/list/?id=46483&type=studio":           {Type: "studio", ID: 46483},
		"https://r18.dev/videos/vod/movies/list/?type=label&id=25637":            {Type: "label", ID: 25637},
		" https://r18.dev/videos/vod/movies/list/?id=307935&type=category ":      {Type: "category", ID: 307935},
		"https://r18.dev/videos/vod/movies/list2/json?id=8591&type=label&page=3": {Type: "label", ID: 8591},
	}
	for in, want := range ok {
		if got, err := R18devSourceFromURL(in); err != nil || got != want {
			t.Errorf("%q: got %v, %v", in, got, err)
		}
	}
	for _, bad := range []string{
		"https://r18.dev/videos/vod/movies/detail/-/id=vrkm01662/",
		"https://r18.dev/videos/vod/movies/list/?id=1069503&type=actress",
		"https://www.dmm.co.jp/digital/videoa/-/list/=/article=maker/id=46483/",
		"https://r18.dev/videos/vod/movies/list/?id=abc&type=studio",
		"",
	} {
		if _, err := R18devSourceFromURL(bad); err == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
}

func TestR18devCheckCustom(t *testing.T) {
	official := []config.R18devScraperConfig{{ID: "kmpvr", Name: "KMPVR"}}
	cfg := config.R18devScraperConfig{
		Name:       " V:one VR ",
		Sources:    []config.R18devSource{{Type: "studio", ID: 46496}},
		Sites:      []string{"gopj", " VOVS", "GOPJ", ""},
		ContentIDs: []string{" "},
		Exclusive:  true,
	}
	if err := R18devCheckCustom(&cfg, official); err != nil {
		t.Fatal(err)
	}
	if cfg.ID != "v-one-vr" || cfg.Company != "V:one VR" || len(cfg.Sites) != 2 || cfg.Sites[0] != "GOPJ" || cfg.ContentIDs != nil || cfg.Exclusive {
		t.Errorf("cleaned up wrong: %+v", cfg)
	}
	for _, bad := range []config.R18devScraperConfig{
		{Name: "KMPVR", Sources: cfg.Sources}, // preset ID
		{Name: "Nothing"},                     // no source
		{Name: "Actress", Sources: []config.R18devSource{{Type: "actress", ID: 1}}},
		{Name: "", Sources: cfg.Sources},
	} {
		b := bad
		if err := R18devCheckCustom(&b, official); err == nil {
			t.Errorf("%+v should be refused", bad)
		}
	}
}
