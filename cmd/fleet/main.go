package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	fleetapiserver "go.datum.net/fleet/internal/apiserver"
	"go.datum.net/fleet/internal/version"
	"go.datum.net/fleet/pkg/generated/openapi"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	apiopenapi "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/cli"
	basecompatibility "k8s.io/component-base/compatibility"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"

	// Register JSON logging format
	_ "k8s.io/component-base/logs/json/register"
)

func init() {
	utilruntime.Must(logsapi.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
	utilfeature.DefaultMutableFeatureGate.Set("LoggingBetaOptions=true")
	utilfeature.DefaultMutableFeatureGate.Set("RemoteRequestHeaderUID=true")
}

func main() {
	cmd := NewFleetCommand()
	code := cli.Run(cmd)
	os.Exit(code)
}

// NewFleetCommand creates the root command.
func NewFleetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet",
		Short: "Datum fleet management service",
		Long: `Fleet manager is an incubating Datum service for managing machine fleets.

No resource API is served until its initial contract is accepted and implemented.`,
	}

	cmd.AddCommand(NewServeCommand())
	cmd.AddCommand(NewVersionCommand())

	return cmd
}

// NewServeCommand creates the serve subcommand that starts the API server.
func NewServeCommand() *cobra.Command {
	options := NewFleetOptions()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		Long:  "Start the fleet API server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(); err != nil {
				return err
			}
			if err := options.Validate(); err != nil {
				return err
			}
			return Run(options, cmd.Context())
		},
	}

	flags := cmd.Flags()
	options.AddFlags(flags)

	// Add logging flags - this includes the -v flag for verbosity
	logsapi.AddFlags(options.Logs, flags)

	return cmd
}

// NewVersionCommand creates the version subcommand to display build information.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Show the version, git commit, and build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			fmt.Printf("Fleet Server\n")
			fmt.Printf("  Version:       %s\n", info.Version)
			fmt.Printf("  Git Commit:    %s\n", info.GitCommit)
			fmt.Printf("  Git Tree:      %s\n", info.GitTreeState)
			fmt.Printf("  Build Date:    %s\n", info.BuildDate)
			fmt.Printf("  Go Version:    %s\n", info.GoVersion)
			fmt.Printf("  Go Compiler:   %s\n", info.Compiler)
			fmt.Printf("  Platform:      %s\n", info.Platform)
		},
	}

	return cmd
}

// FleetOptions contains API server configuration.
type FleetOptions struct {
	RecommendedOptions *options.RecommendedOptions
	Logs               *logsapi.LoggingConfiguration
	// Add your custom options here
}

// NewFleetOptions creates options with default values.
func NewFleetOptions() *FleetOptions {
	o := &FleetOptions{
		RecommendedOptions: options.NewRecommendedOptions(
			"/registry/fleet.datumapis.com",
			fleetapiserver.Codecs.LegacyCodec(fleetapiserver.Scheme.PrioritizedVersionsAllGroups()...),
		),
		Logs: logsapi.NewLoggingConfiguration(),
	}

	// Disable etcd since storage implementation is external
	o.RecommendedOptions.Etcd = nil

	// Disable admission plugins since this server doesn't mutate or validate resources.
	o.RecommendedOptions.Admission = nil

	return o
}

func (o *FleetOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
}

func (o *FleetOptions) Complete() error {
	return nil
}

// Validate ensures required configuration is provided.
func (o *FleetOptions) Validate() error {
	// Add validation as needed
	return nil
}

// Config builds the complete server configuration from options.
func (o *FleetOptions) Config() (*fleetapiserver.Config, error) {
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost", nil, nil); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	genericConfig := genericapiserver.NewRecommendedConfig(fleetapiserver.Codecs)

	// Set effective version to match the Kubernetes version we're built against.
	genericConfig.EffectiveVersion = basecompatibility.NewEffectiveVersionFromString("1.34", "", "")

	namer := apiopenapi.NewDefinitionNamer(fleetapiserver.Scheme)
	genericConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(openapi.GetOpenAPIDefinitions, namer)
	genericConfig.OpenAPIV3Config.Info.Title = "Fleet"
	genericConfig.OpenAPIV3Config.Info.Version = version.Version

	// Configure OpenAPI v2
	genericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(openapi.GetOpenAPIDefinitions, namer)
	genericConfig.OpenAPIConfig.Info.Title = "Fleet"
	genericConfig.OpenAPIConfig.Info.Version = version.Version

	if err := o.RecommendedOptions.ApplyTo(genericConfig); err != nil {
		return nil, fmt.Errorf("failed to apply recommended options: %w", err)
	}

	serverConfig := &fleetapiserver.Config{
		GenericConfig: genericConfig,
		ExtraConfig:   fleetapiserver.ExtraConfig{},
	}

	return serverConfig, nil
}

// Run initializes and starts the server.
func Run(options *FleetOptions, ctx context.Context) error {
	if err := logsapi.ValidateAndApply(options.Logs, utilfeature.DefaultMutableFeatureGate); err != nil {
		return fmt.Errorf("failed to apply logging configuration: %w", err)
	}

	config, err := options.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	defer logs.FlushLogs()

	klog.Info("Starting fleet server")
	klog.Info("Metrics available at https://<server-address>/metrics")
	return server.Run(ctx)
}
