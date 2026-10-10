package scrape

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
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
		sc, why, _ := r18devBuildScene(nil, r18devFixture(t, c.fixture), "r18dev-test", false)
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
		`{"content_id":"abc00001","release_date":"2020-01-01","categories":[{"id":6793},{"id":6003}]}`:       "compilation (category 6003)",
		`{"content_id":"abc00001","release_date":"2020-01-01","categories":[{"id":6793},{"id":6608}]}`:       "compilation (category 6608)",
		`{"content_id":"abc00001","release_date":"2020-01-01","categories":[{"id":6793},{"id":6609}]}`:       "compilation (category 6609)",
		`{"content_id":"ovvr00001","release_date":"2020-01-01","label_id":26286,"categories":[{"id":6793}]}`: "excluded label",
	}
	for body, want := range cases {
		_, why, reject := r18devBuildScene(nil, body, "r18dev-test", false)
		if why != want || !reject {
			t.Errorf("got (%q, reject=%v), want (%q, reject=true) for %s", why, reject, want, body)
		}
	}

	// vrkm01903: KMP "VR BEST", 1136 minutes, no 6003
	vrBest := r18devFixture(t, "vrkm01903")
	if _, why, reject := r18devBuildScene(nil, vrBest, "r18dev-test", false); why != "compilation (series 217739)" || !reject {
		t.Errorf("vrkm01903 crawl: got (%q, reject=%v), want compilation (series 217739)", why, reject)
	}
	// single scene / rescrape lets it through (moved out of the cover check window, no network here)
	vrBest = strings.Replace(vrBest, `"2026-10-23"`, `"2020-01-01"`, 1)
	sc, why, _ := r18devBuildScene(nil, vrBest, "r18dev-test", true)
	if why != "" || sc.SceneID != "vrkm01903" || sc.Title != "VRKM-1903" {
		t.Errorf("vrkm01903 single scene: got (%q, %q, %q)", why, sc.SceneID, sc.Title)
	}
	// excluded labels are still excluded on a single scene scrape
	ovvr := `{"content_id":"ovvr00001","release_date":"2020-01-01","label_id":26286,"categories":[{"id":6793},{"id":6003}]}`
	if _, why, _ := r18devBuildScene(nil, ovvr, "r18dev-test", true); why != "excluded label" {
		t.Errorf("ovvr single scene: got %q, want excluded label", why)
	}
}

func TestR18dActressName(t *testing.T) {
	cases := []struct{ id, img, romaji, want string }{
		{"1090262", "aise_mina.jpg", "Aisemina", "Mina Aise"},
		// trailing number + Kunrei-shiki filename vs Hepburn name_romaji
		{"1092214", "tatibana_hinata2.jpg", "Tachibanahinata", "Hinata Tachibana"},
		{"1092211", "adati_yuri2.jpg", "Adachiyuri", "Yuri Adachi"},
		{"1", "simizu_tisato.jpg", "Shimizuchisato", "Chisato Shimizu"},
		{"2", "hukuda_zyun.jpg", "Fukudajun", "Jun Fukuda"},
		{"3", "matumoto_syouko.jpg", "Matsumotoshouko", "Shouko Matsumoto"},
		// already split, no clean split, no match: untouched
		{"4", "ozaki_erika.jpg", "Erika Ozaki", "Erika Ozaki"},
		{"5", "aika.jpg", "AIKA", "AIKA"},
		{"6", "tanaka_mai.jpg", "Suzukimai", "Suzukimai"},
		// override wins
		{"1112912", "", "Mochinoyomogi", "Yomogi Mochino"},
	}
	for _, c := range cases {
		a := gjson.Parse(`{"id":` + c.id + `,"image_url":"` + c.img + `","name_romaji":"` + c.romaji + `"}`)
		if got := r18dActressName(a); got != c.want {
			t.Errorf("%s / %s: got %q, want %q", c.img, c.romaji, got, c.want)
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

func TestR18devRetryAfter(t *testing.T) {
	res := func(h string) *resty.Response {
		r := &http.Response{Header: http.Header{}}
		if h != "" {
			r.Header.Set("Retry-After", h)
		}
		return &resty.Response{RawResponse: r}
	}
	cases := []struct {
		header string
		try    int
		want   time.Duration
	}{
		{"", 1, 30 * time.Second}, // no header: 30s per try
		{"", 2, 60 * time.Second},
		{"45", 1, 45 * time.Second},
		{"0", 1, 5 * time.Second},    // never less than 5s
		{"3600", 1, 5 * time.Minute}, // never more than 5 minutes
		{"junk", 1, 30 * time.Second},
	}
	for _, c := range cases {
		if got := r18devRetryAfter(res(c.header), c.try); got != c.want {
			t.Errorf("Retry-After %q try %d: got %s, want %s", c.header, c.try, got, c.want)
		}
	}
}

func TestR18devPacerBackOff(t *testing.T) {
	p := &r18devPacer{delay: r18devDelay}
	for i := 0; i < 20; i++ {
		p.backOff(0)
	}
	if p.delay != r18devMaxDelay {
		t.Errorf("delay after many 429s = %s, want %s", p.delay, r18devMaxDelay)
	}
	p.reset(r18devDelay)
	if p.delay != r18devDelay {
		t.Errorf("delay after reset = %s, want %s", p.delay, r18devDelay)
	}
}
