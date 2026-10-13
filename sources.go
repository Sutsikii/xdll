package main

import (
	"context"
	"errors"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// parsePostID takes the URL of a post on X, YouTube or Instagram, or a tweet
// ID, and returns its bookmark ID.
func parsePostID(s string) (string, error) {
	if m := youtubeID.FindStringSubmatch(s); m != nil {
		return "yt-" + m[1], nil
	}
	if m := instagramCode.FindStringSubmatch(s); m != nil {
		return "ig-" + m[1], nil
	}
	return parseTweetID(s)
}

// fetchPost describes the post of a bookmark ID.
func fetchPost(ctx context.Context, id string) (*post, error) {
	if v, ok := strings.CutPrefix(id, "yt-"); ok {
		return fetchYouTube(ctx, v)
	}
	if code, ok := strings.CutPrefix(id, "ig-"); ok {
		return fetchInstagram(ctx, code)
	}
	return fetchTweet(ctx, id)
}

var bookmarkID = regexp.MustCompile(`^(?:[0-9]+|yt-[A-Za-z0-9_-]{11}|ig-[A-Za-z0-9_-]+)$`)

// validID reports whether id is a bookmark ID, safe to use as a directory name.
func validID(id string) bool { return len(id) <= 64 && bookmarkID.MatchString(id) }

// downloadMedia downloads m to dst, merging its sound when it comes apart.
func downloadMedia(ctx context.Context, m remoteMedia, dst string) (int64, error) {
	if m.Audio == "" {
		return download(ctx, m.URL, dst)
	}
	err := downloadMerged(ctx, m, dst)
	if err != nil && m.Fallback != "" && ctx.Err() == nil {
		return download(ctx, m.Fallback, dst)
	}
	if err != nil {
		return 0, err
	}
	st, err := os.Stat(dst)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

func downloadMerged(ctx context.Context, m remoteMedia, dst string) error {
	video, audio := dst+".video.part", dst+".audio.part"
	defer os.Remove(video)
	defer os.Remove(audio)
	if _, err := download(ctx, m.URL, video); err != nil {
		return err
	}
	if _, err := download(ctx, m.Audio, audio); err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(dst), ".merge-"+filepath.Base(dst))
	if err := mergeFragmented(video, audio, tmp); err != nil {
		os.Remove(tmp)
		return errors.Join(errors.New("fusion du son"), err)
	}
	return os.Rename(tmp, dst)
}

// mediaSize reads the size of a downloaded file, for sites that do not give it.
func mediaSize(path, kind string) (w, h int) {
	if kind != "image" {
		w, h, _ = mp4Size(path)
		return w, h
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	c, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return c.Width, c.Height
}
