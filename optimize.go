package main

import (
	"cmp"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// OptimizeStatus is the progress of Optimize.
type OptimizeStatus struct {
	Running bool  `json:"running"`
	Done    int   `json:"done"` // posts
	Total   int   `json:"total"`
	Saved   int64 `json:"saved"`  // bytes
	Failed  int   `json:"failed"` // posts left as they were: gone from the site, or an error
}

type optimizer struct {
	mu     sync.Mutex
	status OptimizeStatus
	cancel context.CancelFunc
}

// Optimize downloads again, in economy quality, the videos downloaded in
// maximum quality, and keeps each new file when it is smaller. Posts gone
// from their site keep their files. It returns when done or canceled.
func (l *Library) Optimize(ctx context.Context) (OptimizeStatus, error) {
	l.opt.mu.Lock()
	if l.opt.status.Running {
		l.opt.mu.Unlock()
		return OptimizeStatus{}, errors.New("déjà en cours")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ids := l.heavyPosts()
	l.opt.status = OptimizeStatus{Running: true, Total: len(ids)}
	l.opt.cancel = cancel
	l.opt.mu.Unlock()

	for _, id := range ids {
		saved, err := l.optimizePost(ctx, id)
		if ctx.Err() != nil {
			break
		}
		l.opt.mu.Lock()
		l.opt.status.Done++
		l.opt.status.Saved += saved
		if err != nil {
			l.opt.status.Failed++
		}
		l.opt.mu.Unlock()
	}

	l.opt.mu.Lock()
	defer l.opt.mu.Unlock()
	l.opt.status.Running = false
	l.opt.cancel = nil
	return l.opt.status, nil
}

// OptimizeStatus returns the progress of Optimize.
func (l *Library) OptimizeStatus() OptimizeStatus {
	l.opt.mu.Lock()
	defer l.opt.mu.Unlock()
	return l.opt.status
}

// CancelOptimize stops Optimize after the post in progress.
func (l *Library) CancelOptimize() {
	l.opt.mu.Lock()
	defer l.opt.mu.Unlock()
	if l.opt.cancel != nil {
		l.opt.cancel()
	}
}

// heavyPosts returns the posts with videos that economy could make lighter.
func (l *Library) heavyPosts() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	var ids []string
	for _, b := range l.data.Bookmarks {
		if heavy(b) {
			ids = append(ids, b.ID)
		}
	}
	return ids
}

func heavy(b Bookmark) bool {
	if b.Source == "instagram" || b.Source == "pinterest" { // a single version
		return false
	}
	for _, m := range b.Media {
		if m.Kind == "video" && !m.Light {
			return true
		}
	}
	return false
}

// optimizePost replaces the videos of a post by their economy version
// when smaller, and returns the bytes saved.
func (l *Library) optimizePost(ctx context.Context, id string) (int64, error) {
	l.mu.Lock()
	i := l.find(id)
	var b Bookmark
	if i >= 0 {
		b = l.data.Bookmarks[i]
		b.Media = append([]Media{}, b.Media...)
	}
	l.mu.Unlock()
	if i < 0 {
		return 0, errors.New("post introuvable")
	}
	p, err := fetchPost(ctx, id, true)
	if err != nil {
		return 0, err
	}
	if len(p.Media) != len(b.Media) {
		return 0, errors.New("le post a changé")
	}

	type change struct {
		i   int
		tmp string // "" when the stored file is the smallest
		m   Media
	}
	var changes []change
	defer func() {
		for _, c := range changes {
			if c.tmp != "" {
				os.Remove(c.tmp)
			}
		}
	}()
	for i, m := range b.Media {
		rm := p.Media[i]
		if m.Kind != "video" || m.Light || rm.Kind != "video" || !rm.Light {
			continue
		}
		tmp := filepath.Join(l.mediaDir(), filepath.FromSlash(m.File)) + ".light"
		n, err := downloadMedia(ctx, rm, tmp)
		if err != nil {
			os.Remove(tmp)
			return 0, err
		}
		c := change{i: i, m: m}
		c.m.Light = true
		if n < m.Size {
			c.tmp = tmp
			c.m.Size, c.m.Width, c.m.Height = n, rm.Width, rm.Height
			if c.m.Width == 0 {
				c.m.Width, c.m.Height = mediaSize(tmp, m.Kind)
			}
		} else {
			os.Remove(tmp)
		}
		changes = append(changes, c)
	}

	var saved int64
	var renameErr error
	_, err = l.update(id, func(b *Bookmark) error {
		b.Media = slices.Clone(b.Media) // snapshots share the old one
		for _, c := range changes {
			if c.i >= len(b.Media) || b.Media[c.i].File != c.m.File {
				continue
			}
			if c.tmp != "" {
				// Fails while another program has the file open.
				if err := os.Rename(c.tmp, filepath.Join(l.mediaDir(), filepath.FromSlash(c.m.File))); err != nil {
					renameErr = err
					continue
				}
				saved += b.Media[c.i].Size - c.m.Size
			}
			b.Media[c.i] = c.m
		}
		return nil
	})
	return saved, cmp.Or(err, renameErr)
}
