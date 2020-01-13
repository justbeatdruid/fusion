package app

import (
	"context"
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/server"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/options"
	"github.com/chinamobile/nlpt/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/util/term"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
)

func NewServerCommand() *cobra.Command {
	serverRunOptions := options.NewServerRunOptions()
	command := &cobra.Command{
		Use: "database",
		Long: `The database middleware receives requests from console,
then creates services to backend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			version.PrintAndExitIfRequested()
			func(flags *pflag.FlagSet) {
				flags.VisitAll(func(flag *pflag.Flag) {
					klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
				})
			}(cmd.Flags())

			if err := Run(serverRunOptions); err != nil {
				klog.Errorf("error run server: %+v", err)
			}
			return nil
		},
	}
	fs := command.Flags()
	namedFlagSets := serverRunOptions.Flags()
	version.AddFlags(namedFlagSets.FlagSet("global"))
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(command.OutOrStdout())
	command.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	command.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
	return command
}

func Run(serverRunOptions *options.ServerRunOptions) error {
	k8sconfig, err := serverRunOptions.Config()
	if err != nil {
		klog.Fatalf("Cannot create kubernetes config: %+v", err)
	}
	genericServer, err := server.NewGenericServer(serverRunOptions, k8sconfig)
	if err != nil {
		klog.Fatalf("Create generic server error: %+v", err)
	}
	func(ctx context.Context) {
		genericServer.Run(server.SetupSignalHandler())
	}(context.TODO())
	panic("unreachable")
}
