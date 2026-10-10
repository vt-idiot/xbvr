package api

import (
	"strings"
	"testing"

	"github.com/xbapps/xbvr/pkg/models"
)

func TestSceneIsJAVR(t *testing.T) {
	cases := []struct {
		scraper, url string
		want         bool
	}{
		{"r18dev-kmpvr", "https://r18.dev/videos/vod/movies/detail/-/id=vrkm01662/", true},
		{"", "https://r18.dev/videos/vod/movies/detail/-/id=84vrkm00139/", true},       // import box
		{"", "https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=84vrkm00139/", true}, // old R18.com
		{"", "https://www.javdatabase.com/movies/vrkm-139/", false},
		{"slr-kmpvr", "https://www.sexlikereal.com/scenes/whatever-12345", false},
		{"kmpvr", "https://r18.dev/oops", false}, // has a non-R18.dev scraper
	}
	for _, c := range cases {
		s := models.Scene{ScraperId: c.scraper, SceneURL: c.url}
		if got := sceneIsJAVR(&s); got != c.want {
			t.Errorf("sceneIsJAVR(%q, %q) = %v, want %v", c.scraper, c.url, got, c.want)
		}
	}
}

func TestSceneIsSplit(t *testing.T) {
	if !sceneIsSplit(&models.Scene{SceneID: "VRKM-139_split1"}) {
		t.Error("VRKM-139_split1 should be a split")
	}
	if !sceneIsSplit(&models.Scene{SceneID: "VRKM-139", Tags: []models.Tag{{Name: "split javr"}}}) {
		t.Error("tagged scene should be a split")
	}
	if sceneIsSplit(&models.Scene{SceneID: "VRKM-139", Tags: []models.Tag{{Name: "javr"}}}) {
		t.Error("VRKM-139 is not a split")
	}
}

func TestValidateSplit(t *testing.T) {
	orig := models.Scene{SceneID: "VRKM-139", SceneURL: "https://r18.dev/videos/vod/movies/detail/-/id=84vrkm00139/"}
	part := func(n int, files ...string) RequestSplitScenePart {
		return RequestSplitScenePart{Number: n, Filenames: files}
	}
	ok := RequestSplitScene{Original: "hide", Parts: []RequestSplitScenePart{part(1, "a1.mp4", "a1.srt"), part(3, "a3.mp4")}, Keep: []string{"a.mp4"}}
	if err := validateSplit(&orig, &ok); err != nil {
		t.Errorf("valid split rejected: %v", err)
	}

	bad := map[string]RequestSplitScene{
		"no parts":              {Original: "hide"},
		"original":              {Original: "keep", Parts: []RequestSplitScenePart{part(1)}},
		"part 0":                {Original: "hide", Parts: []RequestSplitScenePart{part(0)}},
		"same part twice":       {Original: "hide", Parts: []RequestSplitScenePart{part(1), part(1)}},
		"file in 2 parts":       {Original: "hide", Parts: []RequestSplitScenePart{part(1, "a.mp4"), part(2, "a.mp4")}},
		"file in part and keep": {Original: "delete", Parts: []RequestSplitScenePart{part(1, "a.mp4")}, Keep: []string{"a.mp4"}},
	}
	for name, r := range bad {
		r := r
		if err := validateSplit(&orig, &r); err == nil {
			t.Errorf("%s: accepted, want an error", name)
		}
	}

	split := models.Scene{SceneID: "VRKM-139_split2", SceneURL: orig.SceneURL}
	if err := validateSplit(&split, &ok); err == nil || !strings.Contains(err.Error(), "already a split") {
		t.Errorf("splitting a split: got %v", err)
	}
	slr := models.Scene{SceneID: "slr-123", ScraperId: "slr-kmpvr", SceneURL: "https://www.sexlikereal.com/scenes/x-123"}
	if err := validateSplit(&slr, &ok); err == nil {
		t.Error("non-JAV scene accepted")
	}
}

func TestSplitCopiesAction(t *testing.T) {
	for col, want := range map[string]bool{
		"tags": true, "cast": true, "synopsis": true, "cover_url": true, "release_date_text": true,
		"filenames_arr": false, "is_multipart": false, "title": false, "duration": false,
	} {
		if got := splitCopiesAction(col); got != want {
			t.Errorf("splitCopiesAction(%q) = %v, want %v", col, got, want)
		}
	}
}

func TestSplitSceneID(t *testing.T) {
	if got := splitSceneID("VRKM-139", 3); got != "VRKM-139_split3" {
		t.Errorf("got %s", got)
	}
	if got := splitSceneID("vrkm01662", 1); got != "vrkm01662_split1" {
		t.Errorf("got %s", got)
	}
}
