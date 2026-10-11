package scrape

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestR18devPreviewNewest(t *testing.T) {
	defer func(d time.Duration) { r18devPace.delay = d }(r18devPace.delay)
	r18devPace.delay = 0
	future := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	page := `{"total_results":1519,"results":[
		{"content_id":"h_1116cafr00600","dvd_id":null,"release_date":"` + future + ` 10:00:00","maker_name_en":"CASANOVA","title_ja":"preorder","jacket_thumb_url":"a"},
		{"content_id":"cafuku00024","dvd_id":null,"release_date":"2025-01-24 10:00:00","title_ja":"【VR】複数プレイ福袋","jacket_thumb_url":"b"},
		{"content_id":"ajvrbest00004","dvd_id":null,"release_date":"2024-01-01 10:00:00","title_ja":"x","jacket_thumb_url":"c"},
		{"content_id":"h_1116cafr00554","dvd_id":"CAFR-554","release_date":"2023-06-23 10:00:00","title_ja":"【VR】real one","jacket_thumb_url":"d"},
		{"content_id":"h_1116cafr00553","dvd_id":"CAFR-553","release_date":"2023-06-16 10:00:00","title_ja":"older","jacket_thumb_url":"e"}]}`
	c := newR18devClient()
	c.http.SetTransport(r18devFakeLists(func(r *http.Request) *http.Response {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(page)), Request: r}
	}))
	p, err := r18devPreview(c, config.R18devSource{Type: "studio", ID: 46483})
	if err != nil {
		t.Fatal(err)
	}
	if p.Newest.DisplayID != "CAFR-554" || p.Newest.Thumb != "d" || p.Maker != "CASANOVA" || p.Total != 1519 || strings.Join(p.Prefixes, ",") != "CAFR" {
		t.Errorf("got %+v", p)
	}
}

func TestR18devLooksLikeBestOf(t *testing.T) {
	for title, want := range map[string]bool{
		"【VR】複数プレイ福袋": true,
		"【VR秋のギフト】収録分数1300分超！厳選20作品HQ画質をノーカット収録！CASANOVA 秋の大収穫祭スペシャル！": true,
		"【VR】ハンラシンTHE BEST VR 淫語と接写で射精に導く絶頂のオナ指示109分！":                 true,
		"【VR】SODVR1000作品突破記念スペシャル第3弾！！絶倫で欲求不満の団地妻たち":                   false,
		"【VR】天使もえ 最初で最後のVR 1本限りだから全てを詰め込んだ144分":                        false,
		"【VR】2作品収録！！オナニーを手伝ってあげるスペシャル":                                 false,
	} {
		if got := r18devLooksLikeBestOf(title); got != want {
			t.Errorf("%s: got %v", title, got)
		}
	}
}
