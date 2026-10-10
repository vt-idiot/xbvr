package scrape

import (
	"os"
	"strings"
	"testing"
)

func r18devFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/r18dev/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestR18devBuildScene(t *testing.T) {
	// All fixtures are older than the cover check window, so no network access happens here.
	cases := []struct {
		fixture, sceneID, site, studio, titlePrefix string
		cast                                        []string
	}{
		// has a dvd_id (R18.com era)
		{"84vrkm00139", "VRKM-139", "VRKM", "K M Produce", "VRKM-139 (VR) Ceiling Angle VR", []string{"Kanna Misaki"}},
		// no dvd_id: raw content ID as scene ID, machine translated title -> display ID only
		{"vrkm01662", "vrkm01662", "VRKM", "K M Produce", "VRKM-1662", []string{"Mina Aise"}},
		// reissue: suffix stripped from the scene ID
		{"hnvr00144re01", "hnvr00144", "HNVR", "Hon Naka", "HNVR-144", []string{"Riho Fujimori"}},
		// label "Outer Space Variety VR", Site still the DVD prefix
		{"84exvr00149", "EXVR-149", "EXVR", "K M Produce", "EXVR-149 Slutty", []string{"Haruka Namiki"}},
		// h_ maker code, null label_name_en
		{"h_1711devr00001", "h_1711devr00001", "DEVR", "DOC", "DEVR-001", []string{"Sumire Kuramoto"}},
	}
	for _, c := range cases {
		sc, why, _ := r18devBuildScene(nil, r18devFixture(t, c.fixture), "r18dev-test")
		if why != "" {
			t.Errorf("%s: skipped: %s", c.fixture, why)
			continue
		}
		if sc.SceneID != c.sceneID || sc.SiteID != c.sceneID {
			t.Errorf("%s: SceneID %q SiteID %q, want %q", c.fixture, sc.SceneID, sc.SiteID, c.sceneID)
		}
		if sc.Site != c.site {
			t.Errorf("%s: Site %q, want %q", c.fixture, sc.Site, c.site)
		}
		if sc.Studio != c.studio {
			t.Errorf("%s: Studio %q, want %q", c.fixture, sc.Studio, c.studio)
		}
		if !strings.HasPrefix(sc.Title, c.titlePrefix) {
			t.Errorf("%s: Title %q, want prefix %q", c.fixture, sc.Title, c.titlePrefix)
		}
		if sc.ScraperID != "r18dev-test" {
			t.Errorf("%s: ScraperID %q", c.fixture, sc.ScraperID)
		}
		if strings.Join(sc.Cast, "|") != strings.Join(c.cast, "|") {
			t.Errorf("%s: Cast %v, want %v", c.fixture, sc.Cast, c.cast)
		}
		if len(sc.Covers) == 0 || !strings.Contains(sc.Covers[0], "pl.jpg") {
			t.Errorf("%s: Covers %v", c.fixture, sc.Covers)
		}
		if !strings.HasPrefix(sc.HomepageURL, "https://r18.dev/videos/vod/movies/detail/-/id=") {
			t.Errorf("%s: HomepageURL %q", c.fixture, sc.HomepageURL)
		}
	}
}

func TestR18devBuildSceneRejects(t *testing.T) {
	cases := map[string]string{
		`{"content_id":"abc00001","release_date":"2020-01-01","categories":[{"id":2001}]}`:                   "not VR",
		`{"content_id":"abc00001","release_date":"2020-01-01","categories":[{"id":6793},{"id":6003}]}`:       "compilation",
		`{"content_id":"ovvr00001","release_date":"2020-01-01","label_id":26286,"categories":[{"id":6793}]}`: "excluded label",
	}
	for body, want := range cases {
		_, why, reject := r18devBuildScene(nil, body, "r18dev-test")
		if why != want || !reject {
			t.Errorf("got (%q, reject=%v), want (%q, reject=true) for %s", why, reject, want, body)
		}
	}
}

func TestR18devContentIDFromCover(t *testing.T) {
	cases := map[string]string{
		"https://pics.dmm.co.jp/digital/video/84vrkm00139/84vrkm00139pl.jpg":     "84vrkm00139",
		"https://awsimgsrc.dmm.com/dig/digital/video/vrkm01942/vrkm01942pl.jpg":  "vrkm01942",
		"https://pics.dmm.co.jp/digital/video/hnvr00144re01/hnvr00144re01pl.jpg": "hnvr00144",
		"https://www.sexlikereal.com/s/images/content/sexlikereal.png":           "",
	}
	for in, want := range cases {
		if got := r18devContentIDFromCover(in); got != want {
			t.Errorf("r18devContentIDFromCover(%q) = %q, want %q", in, got, want)
		}
	}
}
