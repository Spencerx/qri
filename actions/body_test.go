package actions

import (
	"bytes"
	"testing"

	"github.com/qri-io/dataset"
	"github.com/qri-io/qri/base"
)

func TestGetBody(t *testing.T) {
	node := newTestNode(t)
	ref := addCitiesDataset(t, node)

	// ds, err := dsfs.LoadDataset(node.Repo.Store(), ref.Path)
	// if err != nil {
	// 	t.Error(err.Error())
	// }
	ds, err := base.ReadDatasetPath(node.Repo, ref.String())
	if err != nil {
		t.Fatal(err)
	}

	data, err := GetBody(node, ds, dataset.JSONDataFormat, nil, 1, 1, false)
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(data, []byte(`[["new york",8500000,44.4,true]]`)) {
		t.Errorf("byte response mismatch. got: %s", string(data))
	}

	if ds.BodyPath != "/map/QmcCcPTqmckdXLBwPQXxfyW2BbFcUT6gqv9oGeWDkrNTyD" {
		t.Errorf("bodypath mismatch")
	}
}
