package main

import (
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	cases := []struct {
		before *Playlist
		want   *Playlist
	}{
		{
			InitPlaylist(),
			&Playlist{
				[]*PlaylistItem{},
				-1,
			},
		},
	}

	for _, c := range cases {
		if !reflect.DeepEqual(c.before, c.want) {
			t.Errorf("TestInit: %q != %q", c.before, c.want)
		}
	}
}

func TestEnqueue(t *testing.T) {
	cases := []struct {
		before      *Playlist
		item        *PlaylistItem
		index       int
		want        *Playlist
		shoulderror bool
	}{
		{
			InitPlaylist(),
			&PlaylistItem{"/Music/theballadofbilbobaggins.mp3", "aaa", true},
			0,
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"/Music/theballadofbilbobaggins.mp3", "aaa", true},
				},
				-1,
			},
			false,
		},
		// Test invalid index
		{
			InitPlaylist(),
			&PlaylistItem{"/Music/iamlordeyayaya.wav", "aaa", true},
			1,
			&Playlist{
				[]*PlaylistItem{},
				-1,
			},
			true,
		},
		// Test hash collision
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
				},
				-1,
			},
			&PlaylistItem{"I too am lorde", "aaa", true},
			1,
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
				},
				-1,
			},
			true,
		},
		// Test selection adjustment
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"iamlorde.m4a", "ya", true},
				},
				0,
			},
			&PlaylistItem{"iamsparticus.flac", "hurr", true},
			0,
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"iamsparticus.flac", "hurr", true},
					&PlaylistItem{"iamlorde.m4a", "ya", true},
				},
				1, // Selection should have been adjusted, we enqueued before the selection
			},
			false,
		},
	}

	for caseno, c := range cases {
		_, err := c.before.Enqueue(c.index, c.item)
		if c.shoulderror != (err != nil) {
			if err != nil {
				t.Errorf("TestEnqueue: case %d returned err when should be nil(%s)", caseno, err.Error())
			} else {
				t.Errorf("TestEnqueue: case %d returned nil when should be err", caseno)
			}
		}
		if !reflect.DeepEqual(c.before, c.want) {
			t.Errorf("TestEnqueue: %q != %q", c.before, c.want)
		}
	}
}

func TestDequeue(t *testing.T) {
	cases := []struct {
		before      *Playlist
		index       int
		hash        string
		want        *Playlist
		shoulderror bool
	}{
		// Test dequeue. NB, selection should reset to -1
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"darude - sandstorm.avi", "a1", true},
				},
				0,
			},
			0,
			"a1",
			&Playlist{
				[]*PlaylistItem{},
				-1,
			},
			false,
		},
		// Test dequeue empty
		{
			InitPlaylist(),
			0,
			"yayaya",
			&Playlist{
				[]*PlaylistItem{},
				-1,
			},
			true,
		},
		// Test mismatching index and hash
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			0,
			"b2",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			true,
		},
		// Test invalid index
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			1337,
			"b2",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			true,
		},
		// Test invalid hash
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			0,
			"c3",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				-1,
			},
			true,
		},
		// Test selection adjustment
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"a_walk_in_the_black_forest.ogg", "a1", true},
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				1,
			},
			0,
			"a1",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"cactus_in_my_yfronts.mid", "b2", true},
				},
				0,
			},
			false,
		},
	}

	for caseno, c := range cases {
		_, _, err := c.before.Dequeue(c.index, c.hash)
		if c.shoulderror != (err != nil) {
			if err != nil {
				t.Errorf("TestEnqueue: case %d returned err when should be nil(%s)", caseno, err.Error())
			} else {
				t.Errorf("TestEnqueue: case %d returned nil when should be err", caseno)
			}
		}
		if !reflect.DeepEqual(c.before, c.want) {
			t.Errorf("TestDequeue: (case %d) %q != %q", caseno, c.before, c.want)
		}
	}
}

func TestSelect(t *testing.T) {
	cases := []struct {
		before      *Playlist
		index       int
		hash        string
		want        *Playlist
		shoulderror bool
	}{
		// Test select
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"airhorn.aac", "a1", true},
				},
				-1,
			},
			0,
			"a1",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"airhorn.aac", "a1", true},
				},
				0,
			},
			false,
		},
		// Test invalid select
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"airhorn.aac", "a1", true},
				},
				-1,
			},
			69,
			"lol",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"airhorn.aac", "a1", true},
				},
				-1,
			},
			true,
		},
		// Test invalid hash
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"illuminati.aiff", "hl3", true},
				},
				-1,
			},
			0,
			"notreally",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"illuminati.aiff", "hl3", true},
				},
				-1,
			},
			true,
		},
		// Test invalid index
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"harderbetterfastergaben.opus", "pootis", true},
				},
				-1,
			},
			3,
			"pootis",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"harderbetterfastergaben.opus", "pootis", true},
				},
				-1,
			},
			true,
		},
		// Test error on selecting text item
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"Half life 3", "hl3", false},
				},
				-1,
			},
			0,
			"hl3",
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"Half life 3", "hl3", false},
				},
				-1,
			},
			true,
		},
	}

	for caseno, c := range cases {
		curindex, curhash, err := c.before.Select(c.index, c.hash)
		if c.shoulderror != (err != nil) { // If err value is unexpected
			if err != nil {
				t.Errorf("TestSelect: case %d returned err when should be nil(%s)", caseno, err.Error())
			} else {
				t.Errorf("TestSelect: case %d returned nil when should be err", caseno)
			}
		} else if !c.shoulderror {
			if curindex != c.index {
				t.Errorf("TestSelect: returned index does not match requested, and err is nil")
			}
			if curhash != c.hash {
				t.Errorf("TestSelect: returned hash does not match requested, and err is nil")
			}
		}
		if !reflect.DeepEqual(c.before, c.want) {
			t.Errorf("TestSelect: %q != %q", c.before, c.want)
		}
	}
}

func TestAdvance(t *testing.T) {
	cases := []struct {
		before *Playlist
		want   *Playlist
	}{
		// Test advance on empty selection
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				-1,
			},
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				-1,
			},
		},
		// Test advance
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				0,
			},
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				1,
			},
		},
		// Test advance on last item
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				1,
			},
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				1,
			},
		},
		// Test skipping of text items
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"Note to self: play more boney m.", "plzno", false},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				0,
			},
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"Note to self: play more boney m.", "plzno", false},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
				},
				2,
			},
		},
		// Test skipping of text items at end of playlist
		{
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
					&PlaylistItem{"Note to self: play more boney m.", "plzno", false},
					&PlaylistItem{"Thomas dolby 4 lyf", "science", false},
				},
				1,
			},
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"rasputin.mp3", "aaa", true},
					&PlaylistItem{"mabaker.mp3", "bbb", true},
					&PlaylistItem{"Note to self: play more boney m.", "plzno", false},
					&PlaylistItem{"Thomas dolby 4 lyf", "science", false},
				},
				1,
			},
		},
	}

	for _, c := range cases {
		got := c.before
		got.Advance()
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("TestAdvance: %q.Advance() == %q, want %q", c.before, got, c.want)
		}
	}
}
