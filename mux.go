package main

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// mergeFragmented writes to dst one MP4 with the video track of the
// fragmented MP4 video and the audio track of audio, as YouTube serves
// them separately. Fragments are copied as they are, interleaved by time.
func mergeFragmented(video, audio, dst string) error {
	vf, err := os.Open(video)
	if err != nil {
		return err
	}
	defer vf.Close()
	af, err := os.Open(audio)
	if err != nil {
		return err
	}
	defer af.Close()
	v, err := readFragmented(vf)
	if err != nil {
		return fmt.Errorf("vidéo : %w", err)
	}
	a, err := readFragmented(af)
	if err != nil {
		return fmt.Errorf("audio : %w", err)
	}

	// The video is track 1, the audio track 2.
	for id, r := range []*fragmented{v, a} {
		r.trackID = uint32(id + 1)
		if err := setTrackID(r.trak, "tkhd", r.trackID); err != nil {
			return err
		}
		if err := setTrackID(r.trex, "", r.trackID); err != nil {
			return err
		}
	}
	mvhd := slices.Clone(v.mvhd)
	putNextTrackID(mvhd, 3)

	mvex := box("mvex", v.mehd, v.trex, a.trex)
	moov := box("moov", mvhd, v.trak, a.trak, mvex)
	ftyp := box("ftyp", []byte("isom"), u32(512), []byte("isomiso6avc1mp41"))

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := out.Write(append(ftyp, moov...)); err != nil {
		out.Close()
		return err
	}
	type next struct {
		f   *os.File
		fr  fragment
		tid uint32
	}
	var frs []next
	for _, fr := range v.fragments {
		frs = append(frs, next{vf, fr, v.trackID})
	}
	for _, fr := range a.fragments {
		frs = append(frs, next{af, fr, a.trackID})
	}
	slices.SortStableFunc(frs, func(x, y next) int { return cmp.Compare(x.fr.start, y.fr.start) })
	for i, n := range frs {
		moof := make([]byte, n.fr.moofSize)
		if _, err := n.f.ReadAt(moof, n.fr.offset); err == nil {
			err = patchMoof(moof, uint32(i+1), n.tid)
		}
		if err == nil {
			_, err = out.Write(moof)
		}
		if err == nil {
			_, err = io.Copy(out, io.NewSectionReader(n.f, n.fr.offset+n.fr.moofSize, n.fr.size-n.fr.moofSize))
		}
		if err != nil {
			out.Close()
			return err
		}
	}
	return out.Close()
}

// fragmented is a single track fragmented MP4: its header boxes, whole, and
// where its fragments are.
type fragmented struct {
	mvhd, mehd, trak, trex []byte
	timescale              uint32
	trackID                uint32 // in the merged file
	fragments              []fragment
}

// fragment is a moof box and the mdat boxes after it.
type fragment struct {
	offset, moofSize, size int64
	start                  float64 // decode time of the first sample, in seconds
}

func readFragmented(f *os.File) (*fragmented, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var r fragmented
	for off := int64(0); off < st.Size(); {
		typ, size, hdr, err := boxAt(f, off, st.Size())
		if err != nil {
			return nil, err
		}
		switch typ {
		case "moov":
			b := make([]byte, size)
			if _, err := f.ReadAt(b, off); err != nil {
				return nil, err
			}
			if err := r.readMoov(b[hdr:]); err != nil {
				return nil, err
			}
		case "moof":
			b := make([]byte, size)
			if _, err := f.ReadAt(b, off); err != nil {
				return nil, err
			}
			start, err := moofStart(b[hdr:])
			if err != nil {
				return nil, err
			}
			if r.timescale == 0 {
				return nil, errors.New("moof avant moov")
			}
			r.fragments = append(r.fragments, fragment{offset: off, moofSize: size, size: size, start: float64(start) / float64(r.timescale)})
		case "mdat":
			if len(r.fragments) == 0 {
				return nil, errors.New("pas un MP4 fragmenté")
			}
			fr := &r.fragments[len(r.fragments)-1]
			if fr.offset+fr.size != off {
				return nil, errors.New("mdat inattendu")
			}
			fr.size += size
		}
		off += size
	}
	if r.trak == nil || r.trex == nil || len(r.fragments) == 0 {
		return nil, errors.New("pas un MP4 fragmenté")
	}
	return &r, nil
}

func (r *fragmented) readMoov(b []byte) error {
	return eachBox(b, func(typ string, whole, body []byte) error {
		switch typ {
		case "mvhd":
			r.mvhd = whole
		case "trak":
			if r.trak != nil {
				return errors.New("plusieurs pistes")
			}
			r.trak = whole
			mdia := child(body, "mdia")
			mdhd := child(mdia, "mdhd")
			off := 12
			if len(mdhd) > 0 && mdhd[0] == 1 {
				off = 20
			}
			if len(mdhd) < off+4 {
				return errors.New("mdhd invalide")
			}
			r.timescale = binary.BigEndian.Uint32(mdhd[off:])
		case "mvex":
			return eachBox(body, func(typ string, whole, _ []byte) error {
				switch typ {
				case "trex":
					r.trex = whole
				case "mehd":
					r.mehd = whole
				}
				return nil
			})
		}
		return nil
	})
}

// moofStart returns the base media decode time of a moof's track.
func moofStart(moof []byte) (uint64, error) {
	tfdt := child(child(moof, "traf"), "tfdt")
	switch {
	case len(tfdt) >= 12 && tfdt[0] == 1:
		return binary.BigEndian.Uint64(tfdt[4:]), nil
	case len(tfdt) >= 8:
		return uint64(binary.BigEndian.Uint32(tfdt[4:])), nil
	}
	return 0, errors.New("tfdt manquant")
}

// patchMoof renumbers a moof (box header included) and moves it to
// trackID. Its data offsets must be relative to the moof.
func patchMoof(moof []byte, seq, trackID uint32) error {
	body := moof[8:]
	mfhd := child(body, "mfhd")
	traf := child(body, "traf")
	tfhd := child(traf, "tfhd")
	if len(mfhd) < 8 || len(tfhd) < 8 {
		return errors.New("moof invalide")
	}
	if tfhd[3]&1 != 0 { // base-data-offset-present: absolute offsets
		return errors.New("fragment non déplaçable")
	}
	binary.BigEndian.PutUint32(mfhd[4:], seq)
	binary.BigEndian.PutUint32(tfhd[4:], trackID)
	return nil
}

// setTrackID sets the track ID of a trak (through its tkhd) or a trex.
func setTrackID(b []byte, in string, id uint32) error {
	full := b[8:] // version and flags, then fields
	if in != "" {
		full = child(b[8:], in)
	}
	off := 4
	if in == "tkhd" {
		off = 12
		if len(full) > 0 && full[0] == 1 {
			off = 20
		}
	}
	if len(full) < off+4 {
		return errors.New("piste invalide")
	}
	binary.BigEndian.PutUint32(full[off:], id)
	return nil
}

// putNextTrackID sets next_track_ID, the last field of mvhd.
func putNextTrackID(mvhd []byte, id uint32) {
	binary.BigEndian.PutUint32(mvhd[len(mvhd)-4:], id)
}

// boxAt reads the header of the box at off.
func boxAt(f *os.File, off, end int64) (typ string, size, hdr int64, err error) {
	var h [16]byte
	if _, err := f.ReadAt(h[:8], off); err != nil {
		return "", 0, 0, err
	}
	size, hdr = int64(binary.BigEndian.Uint32(h[:])), 8
	switch size {
	case 0:
		size = end - off
	case 1:
		if _, err := f.ReadAt(h[8:16], off+8); err != nil {
			return "", 0, 0, err
		}
		size, hdr = int64(binary.BigEndian.Uint64(h[8:])), 16
	}
	if size < hdr || off+size > end {
		return "", 0, 0, fmt.Errorf("boîte %q tronquée", h[4:8])
	}
	return string(h[4:8]), size, hdr, nil
}

// eachBox calls fn with each box of b, whole and without its header.
func eachBox(b []byte, fn func(typ string, whole, body []byte) error) error {
	for len(b) > 0 {
		if len(b) < 8 {
			return errors.New("boîte tronquée")
		}
		size, hdr := uint64(binary.BigEndian.Uint32(b)), uint64(8)
		if size == 1 {
			if len(b) < 16 {
				return errors.New("boîte tronquée")
			}
			size, hdr = binary.BigEndian.Uint64(b[8:]), 16
		}
		if size == 0 {
			size = uint64(len(b))
		}
		if size < hdr || size > uint64(len(b)) {
			return errors.New("boîte tronquée")
		}
		if err := fn(string(b[4:8]), b[:size], b[hdr:size]); err != nil {
			return err
		}
		b = b[size:]
	}
	return nil
}

// child returns the body of the first box of type typ in b, or nil.
func child(b []byte, typ string) []byte {
	var found []byte
	eachBox(b, func(t string, _, body []byte) error {
		if t == typ {
			found = body
			return io.EOF
		}
		return nil
	})
	return found
}

// box makes a box of the concatenated parts.
func box(typ string, parts ...[]byte) []byte {
	n := 8
	for _, p := range parts {
		n += len(p)
	}
	b := make([]byte, 0, n)
	b = append(b, u32(uint32(n))...)
	b = append(b, typ...)
	for _, p := range parts {
		b = append(b, p...)
	}
	return b
}

func u32(n uint32) []byte { return binary.BigEndian.AppendUint32(nil, n) }

// mp4Size returns the size of the first video track of an MP4.
func mp4Size(path string) (w, h int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return 0, 0, err
	}
	for off := int64(0); off < st.Size(); {
		typ, size, hdr, err := boxAt(f, off, st.Size())
		if err != nil {
			return 0, 0, err
		}
		if typ == "moov" {
			b := make([]byte, size-hdr)
			if _, err := f.ReadAt(b, off+hdr); err != nil {
				return 0, 0, err
			}
			eachBox(b, func(typ string, _, body []byte) error {
				// width and height end tkhd, in 16.16 fixed point.
				if tkhd := child(body, "tkhd"); typ == "trak" && len(tkhd) >= 8 && w == 0 {
					w = int(binary.BigEndian.Uint32(tkhd[len(tkhd)-8:]) >> 16)
					h = int(binary.BigEndian.Uint32(tkhd[len(tkhd)-4:]) >> 16)
				}
				return nil
			})
			return w, h, nil
		}
		off += size
	}
	return 0, 0, errors.New("pas de moov")
}
