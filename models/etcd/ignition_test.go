package etcd

import (
	"context"
	"fmt"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testTemplate(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	id := "1.0.0"
	_, err := d.GetTemplate(context.Background(), "cs", id)
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	err = d.PutTemplate(context.Background(), "cs", id, "data", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutTemplate(context.Background(), "cs", id, "data", map[string]string{})
	if err != sabakan.ErrConflicted {
		t.Fatal(err)
	}

	ign, err := d.GetTemplate(context.Background(), "cs", id)
	if err != nil {
		t.Fatal(err)
	}
	if ign != "data" {
		t.Error("wrong template stored", ign)
	}

	err = d.DeleteTemplate(context.Background(), "cs", id)
	if err != nil {
		t.Fatal(err)
	}

	err = d.DeleteTemplate(context.Background(), "cs", id)
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}
}

func testTemplateMetadata(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	_, err := d.GetTemplateIndex(context.Background(), "cs")
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	err = d.PutTemplate(context.Background(), "cs", "1.0.0", "data", map[string]string{"version": "20181010"})
	if err != nil {
		t.Fatal(err)
	}

	index, err := d.GetTemplateIndex(context.Background(), "cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 1 {
		t.Error("wrong number of templates", len(index))
	}
	if index[0].ID != "1.0.0" {
		t.Error("wrong id", index[0].ID)
	}
	if index[0].Metadata["version"] != "20181010" {
		t.Error("wrong version", index[0].Metadata)
	}

	for i := 0; i < sabakan.MaxIgnitions+10; i++ {
		id := fmt.Sprintf("1.1.%d", i)
		err = d.PutTemplate(context.Background(), "cs", id, "data", map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
	}
	index, err = d.GetTemplateIndex(context.Background(), "cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != sabakan.MaxIgnitions {
		t.Error("wrong number of templates", len(index))
	}
	if index[0].ID != fmt.Sprintf("1.1.%d", 10) {
		t.Error("wrong oldest template", index[0])
	}
	if index[sabakan.MaxIgnitions-1].ID != fmt.Sprintf("1.1.%d", sabakan.MaxIgnitions+9) {
		t.Error("wrong latest template", index[sabakan.MaxIgnitions-1])
	}
}

func TestIgnitionTemplate(t *testing.T) {
	t.Run("Template", testTemplate)
	t.Run("TemplateIDs", testTemplateMetadata)
}
