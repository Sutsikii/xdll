package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// A pin is at pinterest.<any domain>/pin/<id>/, sometimes with a title
// before the ID ("easy-bread--123/"). pin.it links redirect there.
var (
	pinterestID = regexp.MustCompile(`(?i)pinterest\.[a-z.]+/pin/(?:[^/?#]*--)?([0-9]{6,20})`)
	pinItLink   = regexp.MustCompile(`(?i)^\s*(?:https?://)?pin\.it/[A-Za-z0-9]+`)
)

// expandPinIt follows a pin.it short link to the pin's URL. Other URLs are
// returned as they are.
func expandPinIt(ctx context.Context, s string) (string, error) {
	if !pinItLink.MatchString(s) {
		return s, nil
	}
	u := strings.TrimSpace(s)
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	res.Body.Close()
	if !pinterestID.MatchString(res.Request.URL.String()) {
		return "", errors.New("lien pin.it vers autre chose qu'une épingle")
	}
	return res.Request.URL.String(), nil
}

// pinImages are the sizes of an image, by name ("orig", "736x"…).
type pinImages map[string]struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// largest returns the URL of the original, or else of the widest size.
func (im pinImages) largest() string {
	for _, k := range []string{"orig", "originals"} {
		if v, ok := im[k]; ok && v.URL != "" {
			return v.URL
		}
	}
	best, w := "", -1
	for _, v := range im {
		if v.URL != "" && v.Width > w {
			best, w = v.URL, v.Width
		}
	}
	return best
}

type pinVideo struct {
	List map[string]struct {
		URL       string `json:"url"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Thumbnail string `json:"thumbnail"`
	} `json:"video_list"`
}

var pinHLS = regexp.MustCompile(`^(https://[^/]+/videos/[a-z]+)/hls/(.+)\.m3u8$`)

// media returns the best MP4 of the video. When Pinterest only lists HLS,
// the 720p MP4 (with its sound) is next to it. Its size is read from the
// file: the list gives the source's.
func (v pinVideo) media() (remoteMedia, bool) {
	rm := remoteMedia{Kind: "video"}
	width := -1
	for _, f := range v.List {
		if rm.Thumb == "" {
			rm.Thumb = f.Thumbnail
		}
		if strings.HasSuffix(f.URL, ".mp4") && f.Width > width {
			rm.URL, width = f.URL, f.Width
		}
	}
	if rm.URL != "" {
		return rm, true
	}
	for _, f := range v.List {
		if m := pinHLS.FindStringSubmatch(f.URL); m != nil {
			rm.URL = m[1] + "/720p/" + m[2] + ".mp4"
			return rm, true
		}
	}
	return rm, false
}

type pinUser struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type pinData struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CreatedAt     string    `json:"created_at"`
	Pinner        *pinUser  `json:"pinner"`
	NativeCreator *pinUser  `json:"native_creator"`
	Images        pinImages `json:"images"`
	Videos        *pinVideo `json:"videos"`
	Story         *struct {
		Pages []struct {
			Blocks []struct {
				Video *pinVideo `json:"video"`
				Image *struct {
					Images pinImages `json:"images"`
				} `json:"image"`
			} `json:"blocks"`
		} `json:"pages"`
	} `json:"story_pin_data"`
	Carousel *struct {
		Slots []struct {
			Images pinImages `json:"images"`
		} `json:"carousel_slots"`
	} `json:"carousel_data"`
}

// fetchPinterest reads a pin from the API of Pinterest's website, which
// needs no account.
func fetchPinterest(ctx context.Context, id string) (*post, error) {
	opts, _ := json.Marshal(map[string]any{"options": map[string]string{"id": id, "field_set_key": "detailed"}})
	u := "https://www.pinterest.com/resource/PinResource/get/?data=" + url.QueryEscape(string(opts))
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Pinterest-PWS-Handler", "www/pin/[id].js")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var body struct {
		Response struct {
			Data *pinData `json:"data"`
		} `json:"resource_response"`
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("épingle introuvable (supprimée ou privée ?)")
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pinterest.com : HTTP %d", res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("pinterest.com : %w", err)
	}
	d := body.Response.Data
	if d == nil {
		return nil, errors.New("épingle introuvable (supprimée ou privée ?)")
	}
	t := &post{
		ID:     "pin-" + id,
		Source: "pinterest",
		URL:    "https://www.pinterest.com/pin/" + id + "/",
		Text:   pinText(d.Title, d.Description),
		Media:  pinMedia(d),
	}
	if a := cmp.Or(d.NativeCreator, d.Pinner); a != nil {
		t.Author, t.AuthorName = a.Username, a.FullName
	}
	t.Created, _ = time.Parse(time.RFC1123Z, d.CreatedAt)
	if len(t.Media) == 0 {
		return nil, errors.New("pinterest.com : épingle sans image ni vidéo")
	}
	return t, nil
}

// pinMedia lists the video of a pin, the pages of an idea pin, the images
// of a carousel, or else its image.
func pinMedia(d *pinData) []remoteMedia {
	cover := d.Images.largest()
	if d.Videos != nil {
		if rm, ok := d.Videos.media(); ok {
			rm.Thumb = cmpOr(cover, rm.Thumb)
			return []remoteMedia{rm}
		}
	}
	var out []remoteMedia
	if d.Story != nil {
		for _, p := range d.Story.Pages {
			for _, b := range p.Blocks {
				if b.Video != nil {
					if rm, ok := b.Video.media(); ok {
						out = append(out, rm)
					}
				} else if b.Image != nil {
					if u := b.Image.Images.largest(); u != "" {
						out = append(out, remoteMedia{Kind: "image", URL: u})
					}
				}
			}
		}
	}
	if len(out) == 0 && d.Carousel != nil {
		for _, s := range d.Carousel.Slots {
			if u := s.Images.largest(); u != "" {
				out = append(out, remoteMedia{Kind: "image", URL: u})
			}
		}
	}
	if len(out) == 0 && cover != "" {
		out = append(out, remoteMedia{Kind: "image", URL: cover})
	}
	if len(out) > maxCarousel {
		out = out[:maxCarousel]
	}
	return out
}

// pinText joins the title and the description, which often repeats it.
func pinText(title, desc string) string {
	title, desc = strings.TrimSpace(title), strings.TrimSpace(desc)
	if title == "" || strings.Contains(desc, title) {
		return desc
	}
	if desc == "" {
		return title
	}
	return title + "\n\n" + desc
}
