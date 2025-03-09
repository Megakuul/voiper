package app

import (
	"github.com/megakuul/voiper/cmd/voiper/flags"
	"github.com/megakuul/voiper/internal/version"
	"github.com/megakuul/voiper/web"
	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func NewRootCmd() *cobra.Command {
	options := NewRootOptions(flags.NewGlobalFlags())

	var rootCmd = &cobra.Command{
		Use:          "voiper",
		Short:        "Voiper Softphone",
		Version:      version.Version(),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	options.globalFlags.Attach(rootCmd.PersistentFlags())

	return rootCmd
}

type RootOptions struct {
	globalFlags *flags.GlobalFlags
}

func NewRootOptions(gFlags *flags.GlobalFlags) *RootOptions {
	return &RootOptions{
		globalFlags: gFlags,
	}
}

func (r *RootOptions) Run() error {
	app := NewApp(
		WithBase(r.globalFlags.Base),
	)

	return wails.Run(&options.App{
		Title:  "voiper",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: web.Asset,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
}
