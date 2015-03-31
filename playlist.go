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
		items:     []*PlaylistItem{},
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

	// appending on the end is necessary
	if idx, err = pl.resolveIndex(idx, len(pl.items)+1); err != nil {
		return
	}
	pl.insert(idx, item)
	newIdx = idx
	pl.changeSelection(true, newIdx)
	// TODO: confirm has been added at idx?
	return
}

func (pl *Playlist) Dequeue(idx int, hash string) (oldIdx int, oldHash string, err error) {
	if idx, err = pl.resolveIndex(idx, len(pl.items)); err != nil {
		return
	}
	if pl.items[idx].Hash != hash {
		err = fmt.Errorf("Hash does not match")
		return
	}
	oldIdx, oldHash = idx, pl.items[idx].Hash
	pl.remove(idx)
	pl.changeSelection(false, oldIdx)
	return
}

// TODO: Way of deselecting current selection
func (pl *Playlist) Select(idx int, hash string) (curIdx int, curHash string, err error) {
	if idx, err = pl.resolveIndex(idx, len(pl.items)); err != nil {
		return
	}
	if pl.items[idx].Hash != hash {
		err = fmt.Errorf("Hash does not match")
		return
	}
	if !pl.items[idx].IsFile {
		err = fmt.Errorf("Can only select a file")
		return
	}

	pl.selection = idx
	curIdx, curHash = pl.selection, pl.items[idx].Hash
	return
}

func (pl *Playlist) Len() int {
	return len(pl.items)
}

func (pl *Playlist) HasSelection() bool {
	return pl.selection >= 0
}

// Advance selects the next File item in the playlist, if it exists. Returns true if selection changed
func (pl *Playlist) Advance() bool {
	if !pl.HasSelection() { // Don't advance if nothing selected
		return false
	}
	for pl.selection++; pl.selection < len(pl.items); pl.selection++ {
		if pl.items[pl.selection].IsFile {
			return true
		}
	}
	//Dropped off bottom, no more file items in playlist. Select none
	pl.selection = -1
	return true
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

func (pl *Playlist) changeSelection(wasEnqueue bool, index int) {
	// If there's no selection, nothing changes. use a select request to change it
	if !pl.HasSelection() {
		return
	}

	if wasEnqueue {
		if index <= pl.selection {
			pl.selection += 1
		}
	} else {
		if index == pl.selection {
			pl.selection = -1 // Remove selection
		} else if index < pl.selection {
			pl.selection -= 1
		}
	}
}

func (pl *Playlist) resolveIndex(idx int, length int) (resolved int, err error) {
	resolved = idx
	if idx < 0 {
		resolved += length
	}
	if resolved < 0 || resolved >= length {
		// Out of range, in some direction
		err = fmt.Errorf("Index out of range")
	}
	return
}
