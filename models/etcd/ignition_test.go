package etcd

import (
	"context"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testTemplate(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	_, err := d.GetTemplate(context.Background(), "cs", "1111")
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	id, err := d.PutTemplate(context.Background(), "cs", "data")
	if err != nil {
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

func testTemplateIDs(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	_, err := d.GetTemplateIDs(context.Background(), "cs")
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	_, err = d.PutTemplate(context.Background(), "cs", "data")
	if err != nil {
		t.Fatal(err)
	}

	ids, err := d.GetTemplateIDs(context.Background(), "cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Error("wrong number of templates", len(ids))
	}

	for i := 0; i < sabakan.MaxIgnitions+10; i++ {
		_, err = d.PutTemplate(context.Background(), "cs", "data")
		if err != nil {
			t.Fatal(err)
		}
	}
	ids, err = d.GetTemplateIDs(context.Background(), "cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != sabakan.MaxIgnitions {
		t.Error("wrong number of templates", len(ids))
	}

}

func TestIgnitionTemplate(t *testing.T) {
	t.Run("Template", testTemplate)
	t.Run("TemplateIDs", testTemplateIDs)
}
