package operator

import (
	"fmt"
	"os"

	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type AnalyzeOptions struct {
	Kubeconfig  string
	ArchivePath string
}

func NewAnalyzeOptions(streams genericclioptions.IOStreams) *AnalyzeOptions {
	return &AnalyzeOptions{}
}

func NewAnalyzeCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewAnalyzeOptions(streams)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run scylla-operator analyze.",
		Long:  "Run the scylla anallyze to analyze either live cluster or gathered data",
		RunE: func(cmd *cobra.Command, args []string) error {

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

	cmd.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "", o.Kubeconfig, "Kubeconfig")
	cmd.Flags().StringVarP(&o.ArchivePath, "archive-path", "", o.ArchivePath, "Specify the directory path for archive created by must-gather")

	return cmd
}

func (o *AnalyzeOptions) Validate() error {
	var errs []error

	_, err := os.Stat(o.ArchivePath)
	if err != nil {
		errs = append(errs, fmt.Errorf("archive-path is not a valid filepath"))
	}

	if len(o.Kubeconfig) != 0 && len(o.ArchivePath) != 0 {
		errs = append(errs, fmt.Errorf("kubeconfig and archive-path can't both be set"))
	}

	if len(o.Kubeconfig) == 0 && len(o.ArchivePath) == 0 {
		errs = append(errs, fmt.Errorf("kubeconfig and archive-path can't bot be empty"))
	}

	return errors.NewAggregate(errs)
}

func (o *AnalyzeOptions) Complete() error {
	return nil
}

func (o *AnalyzeOptions) Run(streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	klog.Infof("%s run", cmd.Name())
	cliflag.PrintFlags(cmd.Flags())

	return nil
}
