package app

import (
	"context"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/megakuul/voiper/internal/config"
	"github.com/megakuul/voiper/internal/sip"
	"github.com/megakuul/voiper/internal/util"
)

type App struct {
	ctx      context.Context
	basePath string

	clientLock sync.Mutex
	client     *sip.Client

	configLock sync.Mutex
	config     *config.Config
}

type AppOption func(*App)

func NewApp(opts ...AppOption) *App {
	app := &App{
		basePath: "",
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// WithBase adds a base path to the application (used to lookup configs etc).
func WithBase(path string) AppOption {
	return func(a *App) {
		a.basePath = path
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.SetDefault(slog.New(slog.NewJSONHandler(util.NewEventWriter(a.ctx, "log"), &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})))
}

func (a *App) ListConfigs() (map[string]bool, error) {
	slog.Debug("This is debug crapa asddfjas d fölasdlf jalsödfjlasjdf öalsdjflöajdsflaösjdflasf")
	slog.Info("Ich bin eine INFORMATION")
	slog.Warn("Listed Configs asdfa sdf")
	slog.Error("ALARM ALARM ALARM")
	return config.ListConfigs(a.basePath)
}

func (a *App) GetConfig(name, decryptionKey string) (*config.Config, error) {
	return config.LoadConfig(filepath.Join(a.basePath, name), decryptionKey)
}

func (a *App) SetConfig(cfg *config.Config, name, encryptionKey string) error {
	return config.WriteConfig(cfg, filepath.Join(a.basePath, name), encryptionKey)
}

func (a *App) RemoveConfig(name string, encrypted bool) error {
	return config.RemoveConfig(filepath.Join(a.basePath, name), encrypted)
}

func (a *App) EnableConfig(name, decryptionKey string) error {
	cfg, err := config.LoadConfig(filepath.Join(a.basePath, name), decryptionKey)
	if err != nil {
		return err
	}
	a.configLock.Lock()
	a.clientLock.Lock()
	defer a.configLock.Unlock()
	defer a.clientLock.Unlock()
	a.config = cfg
	if a.client != nil {
		a.client.Close()
	}
	a.client = sip.NewClient(a.config)
	return a.client.Register(context.TODO())
}
