package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var youtubeID = regexp.MustCompile(`(?i)(?:youtube\.com/(?:shorts/|live/|embed/|watch\?(?:[^#]*&)?v=)|youtu\.be/)([A-Za-z0-9_-]{11})`)

// ytFormat is a stream of the player API.
type ytFormat struct {
	Itag       int    `json:"itag"`
	URL        string `json:"url"`
	MimeType   string `json:"mimeType"`
	Bitrate    int    `json:"bitrate"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	IsDrc      bool   `json:"isDrc"`
	AudioTrack *struct {
		AudioIsDefault bool `json:"audioIsDefault"`
		IsAutoDubbed   bool `json:"isAutoDubbed"`
	} `json:"audioTrack"`
}

type ytPlayer struct {
	PlayabilityStatus struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"playabilityStatus"`
	StreamingData struct {
		Formats         []ytFormat `json:"formats"`
		AdaptiveFormats []ytFormat `json:"adaptiveFormats"`
	} `json:"streamingData"`
	VideoDetails struct {
		VideoID          string `json:"videoId"`
		Title            string `json:"title"`
		ShortDescription string `json:"shortDescription"`
		Author           string `json:"author"`
		ChannelID        string `json:"channelId"`
		Thumbnail        struct {
			Thumbnails []struct {
				URL   string `json:"url"`
				Width int    `json:"width"`
			} `json:"thumbnails"`
		} `json:"thumbnail"`
	} `json:"videoDetails"`
	Microformat struct {
		Renderer struct {
			PublishDate     string `json:"publishDate"`
			OwnerProfileURL string `json:"ownerProfileUrl"`
		} `json:"playerMicroformatRenderer"`
	} `json:"microformat"`
}

// ytClient is a client of the player API. The Android one gets stream URLs
// that need no signature; the web one has the date and the channel handle.
type ytClient struct {
	name, version, userAgent string
	extra                    map[string]any
}

var (
	ytAndroid = ytClient{"ANDROID", "20.10.38", "com.google.android.youtube/20.10.38 (Linux; U; Android 11) gzip",
		map[string]any{"androidSdkVersion": 30, "osName": "Android", "osVersion": "11"}}
	ytIOS = ytClient{"IOS", "20.10.4", "com.google.ios.youtube/20.10.4 (iPhone16,2; U; CPU iOS 18_3_2 like Mac OS X;)",
		map[string]any{"deviceMake": "Apple", "deviceModel": "iPhone16,2", "osName": "iPhone", "osVersion": "18.3.2.22D82"}}
	ytWeb = ytClient{"WEB", "2.20250101.00.00", userAgent, nil}
)

func ytPlayerRequest(ctx context.Context, c ytClient, id string) (*ytPlayer, error) {
	info := map[string]any{"clientName": c.name, "clientVersion": c.version, "hl": "fr"}
	for k, v := range c.extra {
		info[k] = v
	}
	body, _ := json.Marshal(map[string]any{
		"videoId":        id,
		"context":        map[string]any{"client": info},
		"contentCheckOk": true,
		"racyCheckOk":    true,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.youtube.com/youtubei/v1/player?prettyPrint=false", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube : HTTP %d", res.StatusCode)
	}
	var p ytPlayer
	return &p, json.NewDecoder(res.Body).Decode(&p)
}

// fetchYouTube describes a video or Short: the best H.264 video stream with
// the original audio track, to merge, or the stream with both as a fallback.
// With economy, the video stream is AV1 and at most 1080p.
func fetchYouTube(ctx context.Context, id string, economy bool) (*post, error) {
	var p *ytPlayer
	var errs []string
	for _, c := range []ytClient{ytAndroid, ytIOS} {
		r, err := ytPlayerRequest(ctx, c, id)
		switch {
		case err != nil:
			errs = append(errs, err.Error())
		case r.PlayabilityStatus.Status != "OK":
			errs = append(errs, cmpOr(r.PlayabilityStatus.Reason, r.PlayabilityStatus.Status))
		default:
			p = r
		}
		if p != nil {
			break
		}
	}
	if p == nil {
		return nil, fmt.Errorf("vidéo indisponible (%s)", strings.Join(errs, " ; "))
	}

	d := p.VideoDetails
	t := &post{
		ID:         "yt-" + id,
		Source:     "youtube",
		URL:        "https://www.youtube.com/watch?v=" + id,
		Author:     d.ChannelID,
		AuthorName: d.Author,
		Text:       strings.TrimSpace(d.Title + "\n\n" + d.ShortDescription),
	}
	// The date and the handle are a bonus.
	if w, err := ytPlayerRequest(ctx, ytWeb, id); err == nil {
		m := w.Microformat.Renderer
		t.Created, _ = time.Parse(time.RFC3339, m.PublishDate)
		t.Created = t.Created.UTC()
		if i := strings.LastIndex(m.OwnerProfileURL, "/@"); i >= 0 {
			t.Author = m.OwnerProfileURL[i+2:]
		}
	}

	video, audio, both := pickYouTubeFormats(p.StreamingData.Formats, p.StreamingData.AdaptiveFormats, economy)
	var rm remoteMedia
	switch {
	case video != nil && audio != nil:
		rm = remoteMedia{Kind: "video", URL: video.URL, Audio: audio.URL, Width: video.Width, Height: video.Height}
		if both != nil {
			rm.Fallback = both.URL
		}
	case both != nil:
		rm = remoteMedia{Kind: "video", URL: both.URL, Width: both.Width, Height: both.Height}
	default:
		return nil, errors.New("youtube : aucun flux MP4")
	}
	rm.Light = economy
	rm.Thumb = youTubeThumb(ctx, id, rm.Height > rm.Width, d.Thumbnail.Thumbnails)
	t.Media = []remoteMedia{rm}
	return t, nil
}

// economyYouTube bounds the short side of YouTube videos in economy.
const economyYouTube = 1080

// pickYouTubeFormats returns the best H.264 video stream, or with economy
// the best AV1 one at most economyYouTube, the best AAC stream of the
// original audio track, and the best stream with both. AV1 takes about half
// the space of H.264 for the same picture.
func pickYouTubeFormats(formats, adaptive []ytFormat, economy bool) (video, audio, both *ytFormat) {
	for i := range formats {
		f := &formats[i]
		if f.URL != "" && strings.HasPrefix(f.MimeType, "video/mp4") && (both == nil || f.Height > both.Height) {
			both = f
		}
	}
	audioRank := func(f *ytFormat) int {
		r := f.Bitrate
		if f.AudioTrack == nil || f.AudioTrack.AudioIsDefault {
			r += 1 << 30
		}
		if !f.IsDrc {
			r += 1 << 29
		}
		return r
	}
	better := func(f, than *ytFormat) bool {
		return than == nil || f.Height*f.Width > than.Height*than.Width ||
			(f.Height*f.Width == than.Height*than.Width && f.Bitrate > than.Bitrate)
	}
	var avc, av1 *ytFormat
	for i := range adaptive {
		f := &adaptive[i]
		small := min(f.Width, f.Height) <= economyYouTube
		switch {
		case f.URL == "":
		case strings.HasPrefix(f.MimeType, `video/mp4; codecs="avc1`):
			if better(f, video) {
				video = f
			}
			if small && better(f, avc) {
				avc = f
			}
		case strings.HasPrefix(f.MimeType, `video/mp4; codecs="av01`):
			if small && better(f, av1) {
				av1 = f
			}
		case strings.HasPrefix(f.MimeType, `audio/mp4; codecs="mp4a`):
			if f.AudioTrack != nil && f.AudioTrack.IsAutoDubbed {
				continue
			}
			if audio == nil || audioRank(f) > audioRank(audio) {
				audio = f
			}
		}
	}
	if economy {
		switch {
		case av1 != nil:
			video = av1
		case avc != nil:
			video = avc
		}
	}
	return video, audio, both
}

// youTubeThumb returns a thumbnail URL, vertical for a Short.
func youTubeThumb(ctx context.Context, id string, vertical bool, listed []struct {
	URL   string `json:"url"`
	Width int    `json:"width"`
}) string {
	if vertical {
		u := "https://i.ytimg.com/vi/" + id + "/oardefault.jpg"
		if req, err := http.NewRequestWithContext(ctx, "HEAD", u, nil); err == nil {
			if res, err := client.Do(req); err == nil {
				res.Body.Close()
				if res.StatusCode == http.StatusOK {
					return u
				}
			}
		}
	}
	best := ""
	w := -1
	for _, t := range listed {
		if t.Width > w {
			best, w = t.URL, t.Width
		}
	}
	return best
}
