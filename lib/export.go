package lib

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/qri-io/dataset/dsio"

	"github.com/qri-io/dataset"

	"github.com/qri-io/dataset/dsutil"
	"github.com/qri-io/qri/actions"
	"github.com/qri-io/qri/base"
	"github.com/qri-io/qri/p2p"
	"github.com/qri-io/qri/repo"
)

// ExportRequests encapsulates business logic of export operation
type ExportRequests struct {
	node *p2p.QriNode
	cli  *rpc.Client
}

// CoreRequestsName implements the Requests interface
func (r ExportRequests) CoreRequestsName() string { return "export" }

// NewExportRequests creates a ExportRequests pointer from either a repo
// or an rpc.Client
func NewExportRequests(node *p2p.QriNode, cli *rpc.Client) *ExportRequests {
	if node != nil && cli != nil {
		panic(fmt.Errorf("both node and client supplied to NewExportRequests"))
	}
	return &ExportRequests{
		node: node,
		cli:  cli,
	}
}

// ExportParams defines parameters for the export method
type ExportParams struct {
	Ref     repo.DatasetRef
	RootDir string
	PeerDir bool
	Format  string
}

// Export exports a dataset in the specified format
func (r *ExportRequests) Export(p *ExportParams, ok *bool) error {
	if r.cli != nil {
		return r.cli.Call("ExportRequests.Export", p, ok)
	}

	ref := p.Ref

	// Handle `qri use` to get the current default dataset.
	if err := DefaultSelectedRef(r.node.Repo, &ref); err != nil {
		return err
	}

	ds, err := base.ReadDatasetPath(r.node.Repo, ref.String())
	if err != nil {
		return err
	}
	defer base.CloseDataset(ds)

	// if err := actions.DatasetHead(r.node, &ref); err != nil {
	// 	return err
	// }

	// ds, err := ref.DecodeDataset()
	// if err != nil {
	// 	return err
	// }

	profile, err := r.node.Repo.Profile()
	if err != nil {
		return err
	}

	// TODO (dlong): The -o option, once it is implemened, can be used to calculate `path`.
	path := p.RootDir
	if p.PeerDir {
		peerName := ref.Peername
		if peerName == "me" {
			peerName = profile.Peername
		}
		path = filepath.Join(path, peerName)
	}
	path = filepath.Join(path, ref.Name)

	switch p.Format {
	case "json":
		bufData, err := actions.GetBody(r.node, ds, dataset.JSONDataFormat, nil, 0, 0, true)
		if err != nil {
			return err
		}
		ds.Body = json.RawMessage(bufData)
		ds.Structure = &dataset.Structure{
			Format:   "json",
			Schema:   ds.Structure.Schema,
			Depth:    ds.Structure.Depth,
			ErrCount: ds.Structure.ErrCount,
		}
		// drop any transform stuff
		ds.Transform = nil

		dst, err := os.Create(fmt.Sprintf("%s.json", path))
		if err != nil {
			return err
		}
		if err := json.NewEncoder(dst).Encode(ds); err != nil {
			return err
		}
		*ok = true
		return nil

	case "xlsx":
		f, err := os.Create(fmt.Sprintf("%s.xlsx", path))
		if err != nil {
			return err
		}

		r, err := dsio.NewEntryReader(ds.Structure, ds.BodyFile())
		if err != nil {
			return err
		}

		st := &dataset.Structure{
			Format: "xlsx",
			// FormatConfig: map[string]interface{}{
			// 	"sheetName": "body",
			// },
		}
		w, err := dsio.NewEntryWriter(st, f)
		if err != nil {
			return err
		}

		if err := dsio.Copy(r, w); err != nil {
			return err
		}
		return w.Close()

	default:
		// default to a zip archive
		dst, err := os.Create(fmt.Sprintf("%s.zip", path))
		if err != nil {
			return err
		}
		store := r.node.Repo.Store()
		// TODO (dlong): Use --body-format here to convert the body and ds.Structure.Format, before
		// passing ds to WriteZipArchive.
		if err = dsutil.WriteZipArchive(store, ds, p.Format, ref.String(), dst); err != nil {
			return err
		}
		*ok = true
		// return dst.Close()
	}

	// TODO (dlong): Document the full functionality of export, and restore this code below. Allow
	// non-zip formats like dataset.json with inline body, body.json by itself, outputting to a
	// a directory, along with yaml, and xlsx.
	/*if path != "" {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}

	if !o.NoBody {
		if bodyFormat == "" {
			bodyFormat = ds.Structure.Format.String()
		}

		df, err := dataset.ParseDataFormatString(bodyFormat)
		if err != nil {
			return err
		}

		p := &lib.ReadParams{
			Format: df,
			Path:   ds.Path().String(),
			All:    true,
		}
		r := &lib.ReadResult{}

		if err = o.DatasetRequests.ReadBody(p, r); err != nil {
			return err
		}

		dataPath := filepath.Join(path, fmt.Sprintf("data.%s", bodyFormat))
		dst, err := os.Create(dataPath)
		if err != nil {
			return err
		}

		if p.Format == dataset.CBORDataFormat {
			r.Data = []byte(hex.EncodeToString(r.Data))
		}
		if _, err = dst.Write(r.Data); err != nil {
			return err
		}

		if err = dst.Close(); err != nil {
			return err
		}
		printSuccess(o.Out, "exported data to: %s", dataPath)
	}

	dsPath := filepath.Join(path, dsfs.PackageFileDataset.String())
	var dsBytes []byte

	switch format {
	case "json":
		dsBytes, err = json.MarshalIndent(ds, "", "  ")
		if err != nil {
			return err
		}
	default:
		dsBytes, err = yaml.Marshal(ds)
		if err != nil {
			return err
		}
		dsPath = fmt.Sprintf("%s.yaml", strings.TrimSuffix(dsPath, filepath.Ext(dsPath)))
	}
	if err = ioutil.WriteFile(dsPath, dsBytes, os.ModePerm); err != nil {
		return err
	}

	if ds.Transform != nil && ds.Transform.ScriptPath != "" {
		f, err := o.Repo.Store().Get(datastore.NewKey(ds.Transform.ScriptPath))
		if err != nil {
			return err
		}
		scriptData, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		// TODO - transformations should have default file extensions
		if err = ioutil.WriteFile(filepath.Join(path, "transform.sky"), scriptData, os.ModePerm); err != nil {
			return err
		}
		printSuccess(o.Out, "exported transform script to: %s", filepath.Join(path, "transform.sky"))
	}

	printSuccess(o.Out, "exported dataset.json to: %s", dsPath)

	return nil*/
	return nil
}
