package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var instagramCode = regexp.MustCompile(`(?i)instagram\.com/(?:[A-Za-z0-9_.]+/)?(?:p|reels?|tv)/([A-Za-z0-9_-]{5,})`)

// instaFix are InstaFix servers, used when Instagram gives nothing: they
// give the caption, the author and redirects to each media file.
var instaFix = []string{"https://uuinstagram.com", "https://eeinstagram.com"}

// The servers only describe posts to link preview bots.
const previewBot = "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)"

// Instagram puts the whole post in its page for search engines.
const searchBot = "Googlebot/2.1 (+http://www.google.com/bot.html)"

// maxCarousel bounds the media of a post (Instagram allows 20).
const maxCarousel = 20

func fetchInstagram(ctx context.Context, code string) (*post, error) {
	t, err := fetchInstagramPage(ctx, code)
	if err == nil {
		return t, nil
	}
	errs := []string{err.Error()}
	for _, host := range instaFix {
		t, err := fetchInstaFix(ctx, host, code)
		if err == nil {
			return t, nil
		}
		errs = append(errs, err.Error())
	}
	return nil, fmt.Errorf("post introuvable (%s)", strings.Join(errs, " ; "))
}

func fetchInstaFix(ctx context.Context, host, code string) (*post, error) {
	t := &post{
		ID:     "ig-" + code,
		Source: "instagram",
		URL:    "https://www.instagram.com/p/" + code + "/",
	}
	for n := 1; n <= maxCarousel; n++ {
		meta, err := previewMeta(ctx, host+"/p/"+code+"/"+strconv.Itoa(n))
		if err != nil {
			return nil, err
		}
		var rm remoteMedia
		switch {
		case strings.HasPrefix(meta["og:video"], "/videos/"):
			rm = remoteMedia{Kind: "video", URL: host + meta["og:video"]}
		case strings.HasPrefix(meta["og:image"], "/images/"):
			rm = remoteMedia{Kind: "image", URL: host + meta["og:image"]}
		}
		if rm.URL == "" {
			if n == 1 {
				return nil, fmt.Errorf("%s : %s", host, cmpOr(meta["og:description"], "pas de média"))
			}
			break // past the last media
		}
		rm.Width, _ = strconv.Atoi(meta["og:video:width"])
		rm.Height, _ = strconv.Atoi(meta["og:video:height"])
		t.Media = append(t.Media, rm)
		if n == 1 {
			t.Author = strings.TrimPrefix(meta["twitter:title"], "@")
			t.Text = meta["og:description"]
		}
	}
	if t.Author == "" {
		return nil, errors.New(host + " : réponse inattendue")
	}
	return t, nil
}

// igMedia is a post, or a media of a carousel, in the page of a post.
type igMedia struct {
	Code    string `json:"code"`
	TakenAt int64  `json:"taken_at"`
	Width   int    `json:"original_width"`
	Height  int    `json:"original_height"`
	User    struct {
		Username string `json:"username"`
		FullName string `json:"full_name"`
	} `json:"user"`
	Caption *struct {
		Text string `json:"text"`
	} `json:"caption"`
	Videos []struct {
		URL string `json:"url"`
	} `json:"video_versions"`
	Images struct {
		Candidates []struct {
			URL string `json:"url"`
		} `json:"candidates"`
	} `json:"image_versions2"`
	Carousel []igMedia `json:"carousel_media"`
}

var pageData = regexp.MustCompile(`(?s)<script type="application/json"[^>]*>(.*?)</script>`)

// fetchInstagramPage reads the post from its page on Instagram.
func fetchInstagramPage(ctx context.Context, code string) (*post, error) {
	u := "https://www.instagram.com/p/" + code + "/"
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", searchBot)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instagram.com : HTTP %d", res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	m := findIgMedia(b, code)
	if m == nil {
		return nil, errors.New("instagram.com : post absent de la page")
	}
	t := &post{
		ID:         "ig-" + code,
		Source:     "instagram",
		URL:        u,
		Author:     m.User.Username,
		AuthorName: m.User.FullName,
	}
	if m.Caption != nil {
		t.Text = m.Caption.Text
	}
	if m.TakenAt > 0 {
		t.Created = time.Unix(m.TakenAt, 0)
	}
	items := m.Carousel
	if len(items) == 0 {
		items = []igMedia{*m}
	}
	for _, it := range items {
		rm := remoteMedia{Width: it.Width, Height: it.Height}
		var image string
		if len(it.Images.Candidates) > 0 {
			image = it.Images.Candidates[0].URL // the largest
		}
		if len(it.Videos) > 0 {
			rm.Kind, rm.URL, rm.Thumb = "video", it.Videos[0].URL, image
		} else {
			rm.Kind, rm.URL = "image", image
		}
		if rm.URL == "" {
			return nil, errors.New("instagram.com : média sans fichier")
		}
		t.Media = append(t.Media, rm)
	}
	if t.Author == "" {
		return nil, errors.New("instagram.com : réponse inattendue")
	}
	return t, nil
}

// findIgMedia finds the post in the JSON data of its page.
func findIgMedia(page []byte, code string) *igMedia {
	for _, s := range pageData.FindAllSubmatch(page, -1) {
		if !strings.Contains(string(s[1]), `"xig_polaris_media"`) {
			continue
		}
		var data any
		if json.Unmarshal(s[1], &data) != nil {
			continue
		}
		if m := walkIgMedia(data, code); m != nil {
			return m
		}
	}
	return nil
}

func walkIgMedia(v any, code string) *igMedia {
	switch v := v.(type) {
	case map[string]any:
		if p, ok := v["xig_polaris_media"].(map[string]any); ok && p["code"] == code {
			raw, err := json.Marshal(p["if_not_gated_logged_out"])
			var m igMedia
			if err == nil && json.Unmarshal(raw, &m) == nil && m.Code == code {
				return &m
			}
		}
		for _, c := range v {
			if m := walkIgMedia(c, code); m != nil {
				return m
			}
		}
	case []any:
		for _, c := range v {
			if m := walkIgMedia(c, code); m != nil {
				return m
			}
		}
	}
	return nil
}

var metaTag = regexp.MustCompile(`<meta (?:property|name)="([^"]+)" content="([^"]*)"`)

// previewMeta returns the meta tags a preview bot gets from u.
func previewMeta(ctx context.Context, u string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", previewBot)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s : HTTP %d", u, res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	meta := map[string]string{}
	for _, m := range metaTag.FindAllStringSubmatch(string(b), -1) {
		if _, ok := meta[m[1]]; !ok {
			meta[m[1]] = html.UnescapeString(m[2])
		}
	}
	return meta, nil
}
