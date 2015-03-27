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
			&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
			0,
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
				},
				-1,
			},
			false,
		},
		// Test invalid index
		{
			InitPlaylist(),
			&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
			1,
			&Playlist{
				[]*PlaylistItem{
					&PlaylistItem{"I am lorde ya ya ya", "aaa", true},
				},
				-1,
			},
			true,
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
