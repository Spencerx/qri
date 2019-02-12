package cmd

import (
	"fmt"

	"github.com/qri-io/ioes"
	"github.com/qri-io/qri/lib"
	"github.com/spf13/cobra"
)

// NewBodyCommand creates a new `qri body` cobra command to fetch entries from the body of a dataset
func NewBodyCommand(f Factory, ioStreams ioes.IOStreams) *cobra.Command {
	o := &BodyOptions{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:   "body",
		Short: "Get the body of a dataset",
		Long: `
` + "`qri body`" + ` reads entries from a dataset. Default is 50 entries, starting from the beginning of the body. You can using the ` + "`--limit`" + ` and ` + "`--offset`" + ` flags to iterate through the dataset body.`,
		Example: `  show the first 50 rows of a dataset:
  $ qri body me/dataset_name

  show the next 50 rows of a dataset:
  $ qri body --offset 50 me/dataset_name

  save the body as csv to file
  $ qri body -o new_file.csv -f csv me/dataset_name`,
		Annotations: map[string]string{
			"group": "dataset",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "path to write to, default is stdout")
	cmd.Flags().BoolVarP(&o.All, "all", "a", false, "read all dataset entries (overrides limit, offest)")
	cmd.Flags().StringVarP(&o.Format, "format", "f", "json", "format to export. one of [json,csv,cbor]")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 50, "max number of records to read")
	cmd.Flags().IntVarP(&o.Offset, "offset", "s", 0, "number of records to skip")

	return cmd
}

// BodyOptions encapsulates options for the body command
type BodyOptions struct {
	ioes.IOStreams

	Format string
	Output string
	Limit  int
	Offset int
	All    bool
	Ref    string

	UsingRPC        bool
	DatasetRequests *lib.DatasetRequests
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *BodyOptions) Complete(f Factory, args []string) (err error) {
	// if len(args) > 0 {
	// 	o.Ref = args[0]
	// }
	// o.UsingRPC = f.RPC() != nil
	// o.DatasetRequests, err = f.DatasetRequests()
	return err
}

// Run executes the body command
func (o *BodyOptions) Run() error {
	// TODO (dlong): Delete `body` command in a later change. Update docs.
	return fmt.Errorf("this command has been removed, use `qri get body` instead")
}
