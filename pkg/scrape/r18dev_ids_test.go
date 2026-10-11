package scrape

import "testing"

// Cases come from the full R18.dev "VR Exclusive" category (25,071 scenes, Oct 2026).

func TestR18devNormalizeDVDID(t *testing.T) {
	cases := map[string]string{
		// clean, untouched
		"VRKM-139":   "VRKM-139",
		"EXVR-149":   "EXVR-149",
		"VRKM-1662":  "VRKM-1662",
		"OYCVR-025A": "OYCVR-025A",
		// 4 digits as R18.dev prints them (SOD's 3DSVR has them on the cover), kept
		"3DSVR-0137": "3DSVR-0137",
		"3DSVR-0878": "3DSVR-0878",
		"AVVR-0400":  "AVVR-0400", // not SOD: AVVR's own maker (40380, h_1245)
		// missing dash (old R18.com era)
		"CAFR009": "CAFR-009",
		"DSVR027": "DSVR-027",
		"WVRS004": "WVRS-004",
		"CAPI100": "CAPI-100",
		// over / under padded
		"MIVR-00077": "MIVR-077",
		"MIVR00028":  "MIVR-028",
		"ZVR-01":     "ZVR-001",
		"WVR9-01":    "WVR9-001",
		"WVR9-10":    "WVR9-010",
		"WVR9C-01":   "WVR9C-001",
		// lowercase
		"ftvr-007":   "FTVR-007",
		"tami0003":   "TAMI-0003", // 4 digits stay 4
		"3dsvr-0140": "3DSVR-0140",
		// never letters right after the dash
		"WVR6-D071":     "WVR6D-071",
		"YP-Y003":       "YPY-003",
		"YP-P004":       "YPP-004",
		"Booyah-VR0002": "BOOYAHVR-0002",
		// part letters
		"VRAD801B": "VRAD-801B",
		// junk stays junk
		"WA0208-VR-V01": "WA0208-VRV01",
	}
	for in, want := range cases {
		if got := r18devNormalizeDVDID(in); got != want {
			t.Errorf("r18devNormalizeDVDID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestR18devDisplayID(t *testing.T) {
	cases := map[string]string{
		// plain, maker digits, h_ maker codes
		"vrkm01662":         "VRKM-1662",
		"84vrkm00139":       "VRKM-139",
		"55tmavr00004":      "TMAVR-004",
		"1nhvr00154":        "NHVR-154",
		"24dtvr00004":       "DTVR-004",
		"h_1711devr00001":   "DEVR-001",
		"h_910vrvr00094":    "VRVR-094",
		"h_1116cafr00473":   "CAFR-473",
		"h_1145mivr00081":   "MIVR-081",
		"h_955fsvr00012":    "FSVR-012",
		"h_1285bikmvr00008": "BIKMVR-008",
		"h_1479ypp00005":    "YPP-005",
		"ovvr637":           "OVVR-637",
		"bibivr00176":       "BIBIVR-176",
		// SOD 3D series keeps its 3
		"13dsvr02015":  "3DSVR-2015",
		"13dsvr00878":  "3DSVR-878",
		"13dsivr00001": "3DSIVR-001",
		"1sodsvr00006": "SODSVR-006",
		// reissues
		"hnvr00144re01":       "HNVR-144",
		"84kmvr00799re01":     "KMVR-799",
		"h_1337wvr100001re01": "WVR1-001",
		// "re" inside the prefix is not a reissue
		"5725ore00001": "ORE-001",
		"kore00004":    "KORE-004",
		// more than 5 digits
		"h_1337wvr100001": "WVR1-001",
		"rvr000001":       "RVR-001",
		"84kmvr000165":    "KMVR-165",
		"13dsvr002001":    "3DSVR-2001",
		// VR Research Institute sections
		"h_1337wvr9c00001":   "WVR9C-001",
		"h_1337wvr6d00071":   "WVR6D-071",
		"h_1337wvr00006d086": "WVR6D-086",
		"h_1337wvr90012":     "WVR9-012",
		"h_1337wvr08002":     "WVR8-002",
		"h_1337wvr40001":     "WVR4-001",
		"h_1390wvr600001":    "WVR6-001",
		// other makers use a plain WVR
		"5539wvr00001":   "WVR-001",
		"h_1239wvr00002": "WVR-002",
		// part letters
		"urvrsp00494a":     "URVRSP-494A",
		"oycvr00025c":      "OYCVR-025C",
		"h_1190vrad00801b": "VRAD-801B",
		// not an ID
		"": "",
	}
	for in, want := range cases {
		if got := r18devDisplayID(in); got != want {
			t.Errorf("r18devDisplayID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestR18devSceneID(t *testing.T) {
	cases := []struct{ cid, dvd, want string }{
		{"84vrkm00139", "VRKM-139", "VRKM-139"},
		{"84dpvr00033", "DPVR-033", "DPVR-033"},
		{"h_1145mivr00077", "MIVR-00077", "MIVR-077"},
		{"dsvr00027", "DSVR027", "DSVR-027"},
		// no dvd_id: raw content ID, reissue suffix stripped
		{"vrkm01662", "", "vrkm01662"},
		{"dpvr00033", "", "dpvr00033"},
		{"hnvr00144re01", "", "hnvr00144"},
		{"hnvr00144", "", "hnvr00144"},
		{"13dsvr01546re01", "", "13dsvr01546"},
		{"VRKM01942", "", "vrkm01942"},
		{"vrkm01942", "----", "vrkm01942"},
		// reissue that kept the old dvd_id
		{"84kmvr00799re01", "KMVR-799", "KMVR-799"},
	}
	for _, c := range cases {
		if got := r18devSceneID(c.cid, c.dvd); got != c.want {
			t.Errorf("r18devSceneID(%q, %q) = %q, want %q", c.cid, c.dvd, got, c.want)
		}
	}
}

func TestR18devDisplayPrefix(t *testing.T) {
	cases := map[string]string{
		"VRKM-139":    "VRKM",
		"3DSVR-2015":  "3DSVR",
		"WVR6D-071":   "WVR6D",
		"URVRSP-494A": "URVRSP",
		"vrkm01662":   "vrkm01662",
	}
	for in, want := range cases {
		if got := r18devDisplayPrefix(in); got != want {
			t.Errorf("r18devDisplayPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestR18devContentIDFromURL(t *testing.T) {
	cases := map[string]string{
		"https://r18.dev/videos/vod/movies/detail/-/id=vrkm01662/":             "vrkm01662",
		"https://r18.dev/videos/vod/movies/detail/-/combined=84vrkm00139/json": "84vrkm00139",
		"https://www.dmm.co.jp/en/digital/videoa/-/detail/=/cid=84vrkm00139/":  "84vrkm00139",
		"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=hnvr00144re01/":   "hnvr00144",
		"https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=h_1337wvr100001/": "h_1337wvr100001",
		"https://www.javdatabase.com/movies/vrkm-139/":                         "",
		"https://www.sexlikereal.com/scenes/some-scene-12345":                  "",
	}
	for in, want := range cases {
		if got := r18devContentIDFromURL(in); got != want {
			t.Errorf("r18devContentIDFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}
