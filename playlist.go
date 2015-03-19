package main

import (
	"fmt"
)

type PlaylistItem struct {
	Data   string
	Hash   string
	IsFile bool
}

type Playlist struct {
	items     []*PlaylistItem
	selection int
}

func InitPlaylist() *Playlist {
	pl := &Playlist{
		selection: -1,
	}
	return pl
}

func (pl *Playlist) Enqueue(idx int, item *PlaylistItem) (newIdx int, err error) {
	for _, it := range pl.items {
		if it.Hash == item.Hash {
			err = fmt.Errorf("Hash already exists")
			return
		}
	}

	if idx, err = pl.resolveIndex(idx); err != nil {
		return
	}
	pl.insert(idx, item)
	newIdx = idx
	// TODO: confirm has been added at idx?
	return
}

func (pl *Playlist) Dequeue(idx int, hash string) (oldIdx int, oldHash string, err error) {
	if idx, err = pl.resolveIndex(idx); err != nil {
		return
	}
	if pl.items[idx].Hash != hash {
		err = fmt.Errorf("Hash does not match")
		return
	}
	oldIdx, oldHash = idx, pl.items[idx].Hash
	pl.remove(idx)
	return
}

// TODO: Way of deselecting current selection
func (pl *Playlist) Select(idx int, hash string) (curIdx int, curHash string, err error) {
	if idx, err = pl.resolveIndex(idx); err != nil {
		return
	}
	if pl.items[idx].Hash != hash {
		err = fmt.Errorf("Hash does not match")
		return
	}
	pl.selection = idx
	curIdx, curHash = pl.selection, pl.items[idx].Hash
	return
}

func (pl *Playlist) Len() int {
	return len(pl.items)
}

func (pl *Playlist) insert(i int, item *PlaylistItem) {
	// i must be valid index
	pl.items = append(pl.items, nil)
	copy(pl.items[i+1:], pl.items[i:])
	pl.items[i] = item
}

func (pl *Playlist) remove(i int) {
	// i must be valid index
	pl.items[len(pl.items)-1], pl.items = nil, append(pl.items[:i], pl.items[i+1:]...)
}

func (pl *Playlist) resolveIndex(idx int) (resolved int, err error) {
	if idx < 0 {
		resolved = len(pl.items)
	}
	resolved += idx
	if resolved < 0 || resolved >= len(pl.items) {
		// Out of range, in some direction
		err = fmt.Errorf("Index out of range")
	}
	return
}
