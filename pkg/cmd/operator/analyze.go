package operator

import (
	"errors"
	"fmt"
	"os"

	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/scylladb/scylla-operator/pkg/version"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	analyzeLongDescription = templates.LongDesc(`
		analyze traverses associated resources and diagnoses common issues.

		This command is experimental and subject to change without notice.
	`)
)

type AnalyzeOptions struct {
	genericclioptions.ClientConfig

	ArchivePath string
}

func NewAnalyzeOptions(streams genericclioptions.IOStreams) *AnalyzeOptions {
	return &AnalyzeOptions{
		ClientConfig: genericclioptions.NewClientConfig("scylla-operator-analyze"),
	}
}

func NewAnalyzeCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewAnalyzeOptions(streams)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run scylla-operator analyze.",
		Long:  analyzeLongDescription,
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

	o.AddFlags(cmd)

	return cmd
}

func (o *AnalyzeOptions) AddFlags(cmd *cobra.Command) {
	o.ClientConfig.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.ArchivePath, "archive-path", "", o.ArchivePath, "Path to a compressed must-gather archive or a directory having must-gather structure")
}

func (o *AnalyzeOptions) Validate() error {
	var errs []error

	errs = append(errs, o.ClientConfig.Validate())

	if len(o.ArchivePath) > 0 {
		_, err := os.Stat(o.ArchivePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				errs = append(errs, fmt.Errorf("archive path %q does not exist", o.ArchivePath))
			} else {
				errs = append(errs, fmt.Errorf("can't stat archive path %q", o.ArchivePath))
			}
		}
	}

	if len(o.Kubeconfig) != 0 && len(o.ArchivePath) != 0 {
		errs = append(errs, fmt.Errorf("kubeconfig and archive-path can't both be set"))
	}

	return apierrors.NewAggregate(errs)
}

func (o *AnalyzeOptions) Complete() error {
	err := o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	return nil
}

func (o *AnalyzeOptions) Run(streams genericclioptions.IOStreams, cmd *cobra.Command) error {
	klog.Infof("%s version %s", cmd.Name(), version.Get())
	cliflag.PrintFlags(cmd.Flags())

	return nil
}
