package operator

import(
	"fmt"

	"github.com/spf13/cobra"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"k8s.io/apimachinery/pkg/util/errors"
)

type AnalyzeOptions struct{
	Mode string
	GatherDirPath string
}

func (o *AnalyzeOptions) Validate() error {
	var errs []error

	if len(o.Mode) != 0 && o.Mode != "live" && o.Mode != "gathered"{
		errs = append(errs, fmt.Errorf("Invalid mode. Use either 'live' or 'gathered'" ))
	}

	if o.Mode == "gathered" && len(o.GatherDirPath) == 0 {
		errs = append(errs, fmt.Errorf("dir-path not specified"))
	}

	if len(o.Mode) == 0 && len(o.GatherDirPath) == 0 {
		errs = append(errs, fmt.Errorf("Set flag use=live or specify directory path for gathered data"))
	}

	return errors.NewAggregate(errs)
}

func (o *AnalyzeOptions) Complete() error {
	return nil
}

func (o *AnalyzeOptions) Run(streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	fmt.Printf("Analyze run with options: mode='%s', gatherDirPath='%s'\n", o.Mode, o.GatherDirPath)

	// TODO

	return nil
}

func NewAnalyzeOptions(streams genericclioptions.IOStreams) *AnalyzeOptions {
	return &AnalyzeOptions{}
}

func NewAnalyzeCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewAnalyzeOptions(streams)

	cmd := &cobra.Command{
		Use: "analyze",
		Short: "Run the scylla analyze.",
		Long: "Run the scylla anallyze to analyze either live cluster or gathered data",
		RunE: func(cmd *cobra.Command, args []string) error{

			err := o.Validate()

			if err != nil {
				return err
			}

			err = o.Complete()

			if err != nil {
				return err
			}

			err = o.Run(streams, cmd)

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.Mode, "mode", "", o.Mode, "Specify the mode: live or gathered")
	cmd.Flags().StringVarP(&o.GatherDirPath, "dirPath", "", o.GatherDirPath, "Specify the directory path for gathered data")

	return cmd
}