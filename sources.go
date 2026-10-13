package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// parsePostID takes the URL of a post on X or YouTube, or a tweet ID, and
// returns its bookmark ID.
func parsePostID(s string) (string, error) {
	if m := youtubeID.FindStringSubmatch(s); m != nil {
		return "yt-" + m[1], nil
	}
	return parseTweetID(s)
}

// fetchPost describes the post of a bookmark ID.
func fetchPost(ctx context.Context, id string) (*post, error) {
	if v, ok := strings.CutPrefix(id, "yt-"); ok {
		return fetchYouTube(ctx, v)
	}
	return fetchTweet(ctx, id)
}

var bookmarkID = regexp.MustCompile(`^(?:[0-9]+|yt-[A-Za-z0-9_-]{11})$`)

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
