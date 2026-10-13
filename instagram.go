package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var instagramCode = regexp.MustCompile(`(?i)instagram\.com/(?:[A-Za-z0-9_.]+/)?(?:p|reels?|tv)/([A-Za-z0-9_-]{5,})`)

// instaFix are InstaFix servers: Instagram gives nothing without an account,
// they give the caption, the author and redirects to each media file.
var instaFix = []string{"https://uuinstagram.com", "https://eeinstagram.com"}

// The servers only describe posts to link preview bots.
const previewBot = "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)"

// maxCarousel bounds the media of a post (Instagram allows 20).
const maxCarousel = 20

func fetchInstagram(ctx context.Context, code string) (*post, error) {
	var errs []string
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
