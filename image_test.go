package sabakan

import (
	"encoding/json"
	"testing"
	"time"
)

func testImageIndexAppend(t *testing.T) {
	t.Parallel()

	idx := ImageIndex{}.Append(nil)
	if len(idx) != 1 {
		t.Fatal(`len(idx) != 1`, len(idx))
	}
	if idx[0] != nil {
		t.Fatal(`idx[0] != nil`, idx[0])
	}

	idx = ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
		&Image{ID: "3"},
		&Image{ID: "4"},
	}
	idx = idx.Append(&Image{ID: "5"})
	if len(idx) != MaxImages {
		t.Fatal(`len(idx) != MaxImages`, len(idx))
	}

	if idx[0].ID != "1" {
		t.Error(`idx[0].ID != "1"`, idx[0].ID)
	}
	if idx[4].ID != "5" {
		t.Error(`idx[4].ID != "5"`, idx[4].ID)
	}

	idx = ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
		&Image{ID: "3"},
		&Image{ID: "4"},
		&Image{ID: "5"},
		&Image{ID: "6"},
	}
	idx = idx.Append(&Image{ID: "7"})
	if len(idx) != MaxImages {
		t.Fatal(`len(idx) != MaxImages`, len(idx))
	}

	if idx[0].ID != "3" {
		t.Error(`idx[0].ID != "3"`, idx[0].ID)
	}
	if idx[4].ID != "7" {
		t.Error(`idx[4].ID != "7"`, idx[4].ID)
	}
}

func testImageIndexFind(t *testing.T) {
	t.Parallel()

	idx := ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
		&Image{ID: "3"},
		&Image{ID: "4"},
	}

	img := idx.Find("1")
	if img == nil {
		t.Fatal("Image ID 1 should be found")
	}
	if img.ID != "1" {
		t.Error(`img.ID != "1"`, img.ID)
	}

	img = idx.Find("foo")
	if img != nil {
		t.Error("Image ID foo should not be found")
	}
}

func testImageIndexJSON(t *testing.T) {
	t.Parallel()

	dt := time.Date(2010, 2, 3, 4, 5, 6, 123456, time.UTC)
	idx := ImageIndex{
		&Image{ID: "0", Date: dt},
	}

	j, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(j))

	var idx2 ImageIndex
	err = json.Unmarshal(j, &idx2)
	if err != nil {
		t.Fatal(err)
	}

	if len(idx2) != 1 {
		t.Fatal(`len(idx2) != 1`, len(idx2))
	}
	img := idx2[0]
	if img.ID != "0" {
		t.Error(`img.ID != "0"`, img.ID)
	}
	if !img.Date.Equal(dt) {
		t.Error(`!img.Date.Equal(dt)`, img.Date)
	}
}

func TestImageIndex(t *testing.T) {
	t.Run("Append", testImageIndexAppend)
	t.Run("Find", testImageIndexFind)
	t.Run("JSON", testImageIndexJSON)
}