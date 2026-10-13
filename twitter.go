package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// post is a post as the public APIs describe it.
type post struct {
	ID         string // the bookmark ID: the tweet ID, "yt-" or "ig-" and the site's
	Source     string // "x", "youtube" or "instagram"
	URL        string
	Author     string // screen name or handle, without @
	AuthorName string
	Text       string
	Created    time.Time
	Media      []remoteMedia
}

// remoteMedia is a photo, video or GIF not downloaded yet.
type remoteMedia struct {
	Kind     string // "image", "video" or "gif"
	URL      string // the best quality file
	Audio    string // the sound of URL, to merge, when served apart
	Fallback string // a file with both, when merging fails
	Thumb    string // a poster image, for videos and GIFs
	Width    int
	Height   int
	Light    bool // in economy quality
}

var statusID = regexp.MustCompile(`(?i)/status(?:es)?/(\d+)`)

// parseTweetID takes a post URL or ID.
func parseTweetID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("URL vide")
	}
	if _, err := strconv.ParseUint(s, 10, 64); err == nil {
		return s, nil
	}
	if m := statusID.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	return "", errors.New("URL invalide")
}

var client = &http.Client{Timeout: 30 * time.Second}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"

func getJSON(ctx context.Context, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s : HTTP %d", u, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// fetchTweet tries the syndication API, then FxTwitter. With economy,
// videos are at most 720p.
func fetchTweet(ctx context.Context, id string, economy bool) (*post, error) {
	t, err := fetchSyndication(ctx, id, economy)
	if err == nil && len(t.Media) > 0 {
		return t, nil
	}
	t2, err2 := fetchFx(ctx, id, economy)
	if err2 == nil {
		return t2, nil
	}
	if err == nil {
		return t, nil // a post without media
	}
	return nil, fmt.Errorf("post introuvable (%v ; %v)", err, err2)
}

// syndicationToken is the token embeds send with the ID.
func syndicationToken(id string) string {
	n, _ := strconv.ParseFloat(id, 64)
	x := n / 1e15 * math.Pi
	ip, frac := math.Modf(x)
	s := strconv.FormatInt(int64(ip), 36)
	for i := 0; i < 10 && frac > 0; i++ {
		frac *= 36
		d, f := math.Modf(frac)
		s += strconv.FormatInt(int64(d), 36)
		frac = f
	}
	s = strings.NewReplacer("0", "", ".", "").Replace(s)
	if s == "" {
		s = "a"
	}
	return s
}

func fetchSyndication(ctx context.Context, id string, economy bool) (*post, error) {
	var r struct {
		Typename  string `json:"__typename"`
		IDStr     string `json:"id_str"`
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"user"`
		NoteTweet struct {
			Text string `json:"text"`
		} `json:"note_tweet"`
		MediaDetails []struct {
			Type          string `json:"type"`
			MediaURLHTTPS string `json:"media_url_https"`
			OriginalInfo  struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"original_info"`
			VideoInfo struct {
				Variants []struct {
					Bitrate     int    `json:"bitrate"`
					ContentType string `json:"content_type"`
					URL         string `json:"url"`
				} `json:"variants"`
			} `json:"video_info"`
		} `json:"mediaDetails"`
	}
	u := "https://cdn.syndication.twimg.com/tweet-result?lang=fr&id=" + id + "&token=" + syndicationToken(id)
	if err := getJSON(ctx, u, &r); err != nil {
		return nil, err
	}
	if r.IDStr == "" {
		return nil, fmt.Errorf("syndication : %s", cmpOr(r.Typename, "réponse vide"))
	}
	t := &post{
		ID:         r.IDStr,
		Source:     "x",
		Author:     r.User.ScreenName,
		AuthorName: r.User.Name,
		Text:       cmpOr(r.NoteTweet.Text, r.Text),
	}
	t.Created, _ = time.Parse(time.RFC3339, r.CreatedAt)
	for _, m := range r.MediaDetails {
		rm := remoteMedia{Width: m.OriginalInfo.Width, Height: m.OriginalInfo.Height}
		switch m.Type {
		case "photo":
			rm.Kind = "image"
			rm.URL = originalPhoto(m.MediaURLHTTPS)
		case "video", "animated_gif":
			rm.Kind = "video"
			if m.Type == "animated_gif" {
				rm.Kind = "gif"
			}
			rm.Thumb = m.MediaURLHTTPS
			var vs []xVariant
			for _, v := range m.VideoInfo.Variants {
				if v.ContentType == "video/mp4" {
					vs = append(vs, xVariant{v.URL, v.Bitrate})
				}
			}
			rm.pickVariant(vs, economy)
		}
		if rm.URL != "" {
			t.Media = append(t.Media, rm)
		}
	}
	t.Text = stripMediaLinks(t.Text)
	t.URL = postURL(t)
	return t, nil
}

func fetchFx(ctx context.Context, id string, economy bool) (*post, error) {
	var r struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Tweet   struct {
			ID               string `json:"id"`
			URL              string `json:"url"`
			Text             string `json:"text"`
			CreatedTimestamp int64  `json:"created_timestamp"`
			Author           struct {
				Name       string `json:"name"`
				ScreenName string `json:"screen_name"`
			} `json:"author"`
			Media struct {
				All []struct {
					Type         string `json:"type"`
					URL          string `json:"url"`
					ThumbnailURL string `json:"thumbnail_url"`
					Width        int    `json:"width"`
					Height       int    `json:"height"`
					Formats      []struct {
						Container string `json:"container"`
						Bitrate   int    `json:"bitrate"`
						URL       string `json:"url"`
					} `json:"formats"`
				} `json:"all"`
			} `json:"media"`
		} `json:"tweet"`
	}
	if err := getJSON(ctx, "https://api.fxtwitter.com/status/"+id, &r); err != nil {
		return nil, err
	}
	if r.Code != 200 {
		return nil, fmt.Errorf("fxtwitter : %s", r.Message)
	}
	t := &post{
		ID:         r.Tweet.ID,
		Source:     "x",
		URL:        r.Tweet.URL,
		Author:     r.Tweet.Author.ScreenName,
		AuthorName: r.Tweet.Author.Name,
		Text:       r.Tweet.Text,
		Created:    time.Unix(r.Tweet.CreatedTimestamp, 0).UTC(),
	}
	for _, m := range r.Tweet.Media.All {
		rm := remoteMedia{URL: m.URL, Thumb: m.ThumbnailURL, Width: m.Width, Height: m.Height}
		switch m.Type {
		case "photo":
			rm.Kind = "image"
			rm.URL = originalPhoto(m.URL)
			rm.Thumb = ""
		case "gif":
			rm.Kind = "gif"
		default:
			rm.Kind = "video"
			var vs []xVariant
			for _, f := range m.Formats {
				if f.Container == "mp4" {
					vs = append(vs, xVariant{f.URL, f.Bitrate})
				}
			}
			if len(vs) > 0 {
				rm.pickVariant(vs, economy)
			}
		}
		t.Media = append(t.Media, rm)
	}
	if t.URL == "" {
		t.URL = postURL(t)
	}
	return t, nil
}

// xVariant is one of the sizes of a video, as an MP4.
type xVariant struct {
	URL     string
	Bitrate int
}

// economyX bounds the short side of X videos in economy: X serves 720p
// at a quarter to a third of the size of 1080p.
const economyX = 720

var variantSize = regexp.MustCompile(`/(\d+)x(\d+)/`)

// pickVariant sets m to the best variant, or with economy the best one at
// most economyX, and its size when the URL tells it.
func (m *remoteMedia) pickVariant(vs []xVariant, economy bool) {
	best, bestLight := -1, -1
	for i, v := range vs {
		if best < 0 || v.Bitrate >= vs[best].Bitrate {
			best = i
		}
		if s := variantSize.FindStringSubmatch(v.URL); s != nil {
			w, _ := strconv.Atoi(s[1])
			h, _ := strconv.Atoi(s[2])
			if min(w, h) <= economyX && (bestLight < 0 || v.Bitrate >= vs[bestLight].Bitrate) {
				bestLight = i
			}
		}
	}
	if best < 0 {
		return
	}
	pick := best
	if economy && bestLight >= 0 {
		pick = bestLight
	}
	m.URL, m.Light = vs[pick].URL, economy
	if s := variantSize.FindStringSubmatch(m.URL); s != nil {
		m.Width, _ = strconv.Atoi(s[1])
		m.Height, _ = strconv.Atoi(s[2])
	}
}

// originalPhoto returns the original size URL of a photo.
func originalPhoto(u string) string {
	p, err := url.Parse(u)
	if err != nil || p.Host != "pbs.twimg.com" {
		return u
	}
	ext := strings.TrimPrefix(path.Ext(p.Path), ".")
	if ext == "" {
		ext = cmpOr(p.Query().Get("format"), "jpg")
	}
	p.Path = strings.TrimSuffix(p.Path, path.Ext(p.Path))
	p.RawQuery = url.Values{"format": {ext}, "name": {"orig"}}.Encode()
	return p.String()
}

var tco = regexp.MustCompile(`\s*https://t\.co/\w+\s*$`)

func stripMediaLinks(s string) string { return strings.TrimSpace(tco.ReplaceAllString(s, "")) }

func postURL(t *post) string {
	return "https://x.com/" + cmpOr(t.Author, "i") + "/status/" + t.ID
}

// download writes u to dst through a temporary file.
func download(ctx context.Context, u, dst string) (int64, error) {
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".part-*")
	if err != nil {
		return 0, err
	}
	var n int64
	if p, perr := url.Parse(u); perr == nil && strings.HasSuffix(p.Host, ".googlevideo.com") {
		n, err = getRanges(ctx, u, tmp)
	} else {
		n, err = getRange(ctx, u, -1, tmp)
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), dst)
	}
	if err != nil {
		os.Remove(tmp.Name())
		return 0, err
	}
	return n, nil
}

var dl = &http.Client{} // no timeout for large videos

// getRange writes u to w from byte from, or all of it when from is -1.
func getRange(ctx context.Context, u string, from int64, w io.Writer) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	want := http.StatusOK
	if from >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", from, from+rangeSize-1))
		want = http.StatusPartialContent
	}
	res, err := dl.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return 0, io.EOF
	}
	if res.StatusCode != want {
		return 0, fmt.Errorf("téléchargement de %s : HTTP %d", u, res.StatusCode)
	}
	return io.Copy(w, res.Body)
}

// rangeSize is the size of the parts of a YouTube stream: it slows down
// whole downloads to the playback speed, not parts.
const rangeSize = 10 << 20

func getRanges(ctx context.Context, u string, w io.Writer) (int64, error) {
	var total int64
	for {
		n, err := getRange(ctx, u, total, w)
		total += n
		if err == io.EOF || (err == nil && n < rangeSize) {
			return total, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

// extOf returns the file extension of a media URL.
func extOf(u, kind string) string {
	p, err := url.Parse(u)
	if err == nil {
		if f := p.Query().Get("format"); f != "" {
			return "." + f
		}
		if e := path.Ext(p.Path); e != "" && len(e) <= 5 {
			return strings.ToLower(e)
		}
	}
	if kind == "image" {
		return ".jpg"
	}
	return ".mp4"
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
