package etcd

import (
	"context"
	"testing"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/google/go-cmp/cmp"
)

func testTemplate(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	id := "1.0.0"
	_, err := d.GetTemplate(context.Background(), "cs", id)
	if err != sabakan.ErrNotFound {
		t.Fatal("unexpected error: ", err)
	}

	tmpl := &sabakan.IgnitionTemplate{
		Version: sabakan.Ignition2_3,
	}
	err = d.PutTemplate(context.Background(), "cs", id, tmpl)
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutTemplate(context.Background(), "cs", id, tmpl)
	if err != sabakan.ErrConflicted {
		t.Fatal(err)
	}

	ign, err := d.GetTemplate(context.Background(), "cs", id)
	if err != nil {
		t.Fatal(err)
	}
	if ign.Version != sabakan.Ignition2_3 {
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

	tmpl := &sabakan.IgnitionTemplate{
		Version: sabakan.Ignition2_3,
	}
	err := d.PutTemplate(context.Background(), "cs", "1.2.3", tmpl)
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutTemplate(context.Background(), "cs", "1.4.0", tmpl)
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutTemplate(context.Background(), "cs", "1.2", tmpl)
	if err != nil {
		t.Fatal(err)
	}

	ids, err := d.GetTemplateIDs(context.Background(), "cs")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"1.2", "1.2.3", "1.4.0"}
	if !cmp.Equal(expected, ids) {
		t.Error("wrong template IDs:", cmp.Diff(expected, ids))
	}
}

func TestIgnitionTemplate(t *testing.T) {
	t.Run("Template", testTemplate)
	t.Run("TemplateIDs", testTemplateIDs)
}
