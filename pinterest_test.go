package main

import "testing"

func TestPinVideo(t *testing.T) {
	var v pinVideo
	v.List = map[string]struct {
		URL       string `json:"url"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Thumbnail string `json:"thumbnail"`
	}{"V_HLSV3_MOBILE": {URL: "https://v1.pinimg.com/videos/iht/hls/48/26/c1/abc.m3u8", Thumbnail: "https://i.pinimg.com/t.jpg"}}
	rm, ok := v.media()
	if !ok || rm.URL != "https://v1.pinimg.com/videos/iht/720p/48/26/c1/abc.mp4" || rm.Thumb != "https://i.pinimg.com/t.jpg" {
		t.Errorf("HLS only: %+v %v", rm, ok)
	}
}

func TestPinText(t *testing.T) {
	for _, c := range [][3]string{
		{"Easy bread", "Easy bread . #bread", "Easy bread . #bread"},
		{"Easy bread", "Ready in an hour", "Easy bread\n\nReady in an hour"},
		{" ", " Ready ", "Ready"},
		{"Easy bread", "", "Easy bread"},
	} {
		if got := pinText(c[0], c[1]); got != c[2] {
			t.Errorf("pinText(%q, %q) = %q", c[0], c[1], got)
		}
	}
}
