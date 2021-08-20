package sabakan

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func testImageValidID(t *testing.T) {
	t.Parallel()

	if !IsValidImageID("ubuntu-bionic") {
		t.Error(`!IsValidImageID("ubuntu-bionic")`)
	}
	if !IsValidImageID("Bionic20180503") {
		t.Error(`!IsValidImageID("Bionic20180503")`)
	}
	if !IsValidImageID("123.45.6") {
		t.Error(`!IsValidImageID("123.45.6")`)
	}

	if IsValidImageID("Ubuntu 18.04") {
		t.Error(`IsValidImageID("Ubuntu 18.04")`)
	}
}

func testImageValidOS(t *testing.T) {
	t.Parallel()

	if !IsValidImageOS("coreos") {
		t.Error(`!IsValidImageOS("coreos")`)
	}
	if !IsValidImageOS("ubuntu18.04") {
		t.Error(`!IsValidImageOS("ubuntu18.04")`)
	}

	if IsValidImageOS("ubuntu 18.04") {
		t.Error(`IsValidImageOS("ubuntu 18.04")`)
	}
	if IsValidImageOS("Ubuntu") {
		t.Error(`IsValidImageOS("Ubuntu")`)
	}
}

func testImageValid(t *testing.T) {
	t.Run("ID", testImageValidID)
	t.Run("OS", testImageValidOS)
}

func testImageIndexAppend(t *testing.T) {
	t.Parallel()

	idx, _ := ImageIndex{}.Append(&Image{ID: ""})
	if len(idx) != 1 {
		t.Fatal(`len(idx) != 1`, len(idx))
	}
	if idx[0] == nil {
		t.Fatal(`idx[0] == nil`)
	}
	if idx[0].ID != "" {
		t.Fatal(`idx[0].ID != ""`, idx[0].ID)
	}

	idx = ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
		&Image{ID: "3"},
		&Image{ID: "4"},
	}
	idx, _ = idx.Append(&Image{ID: "5"})
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
	idx, dels := idx.Append(&Image{ID: "7"})
	if len(idx) != MaxImages {
		t.Fatal(`len(idx) != MaxImages`, len(idx))
	}

	if idx[0].ID != "3" {
		t.Error(`idx[0].ID != "3"`, idx[0].ID)
	}
	if idx[4].ID != "7" {
		t.Error(`idx[4].ID != "7"`, idx[4].ID)
	}
	if !reflect.DeepEqual(dels, []string{"0", "1", "2"}) {
		t.Error(`dels != {"0", "1", "2"}`)
	}

	idx = ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1", URLs: []string{"aaa"}},
		&Image{ID: "2"},
	}
	idx, _ = idx.Append(&Image{ID: "1"})
	if len(idx) != 3 {
		t.Fatal(`len(idx) != 3`, len(idx))
	}
	if idx[0].ID != "0" {
		t.Error(`idx[0].ID != "0"`, idx[0].ID)
	}
	if idx[1].ID != "2" {
		t.Error(`idx[1].ID != "2"`, idx[1].ID)
	}
	if idx[2].ID != "1" {
		t.Error(`idx[2].ID != "1"`, idx[2].ID)
	}
	if len(idx[2].URLs) != 1 {
		t.Error(`len(idx[2].URLs) != 1`, idx[2].URLs)
	}

	idx = ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
	}
	idx, _ = idx.Append(&Image{ID: "2"})
	if len(idx) != 3 {
		t.Fatal(`len(idx) != 3`, len(idx))
	}
	if idx[2].ID != "2" {
		t.Error(`idx[2].ID != "2"`, idx[2].ID)
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

func testImageIndexRemove(t *testing.T) {
	t.Parallel()

	idx := ImageIndex{
		&Image{ID: "0"},
		&Image{ID: "1"},
		&Image{ID: "2"},
		&Image{ID: "3"},
		&Image{ID: "4"},
	}

	idx2 := idx.Remove("6")
	if len(idx2) != len(idx) {
		t.Error(`len(idx2) != len(idx)`, len(idx2))
	}
	if !reflect.DeepEqual(idx2, idx) {
		t.Error(`!reflect.DeepEqual(idx2, idx)`)
	}

	idx2 = idx.Remove("2")
	if len(idx2) != (len(idx) - 1) {
		t.Error(`len(idx2) != (len(idx) - 1)`)
	}
	img := idx2.Find("2")
	if img != nil {
		t.Error("Image ID 2 should not be found")
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
	t.Run("Valid", testImageValid)
	t.Run("Append", testImageIndexAppend)
	t.Run("Find", testImageIndexFind)
	t.Run("Remove", testImageIndexRemove)
	t.Run("JSON", testImageIndexJSON)
}
