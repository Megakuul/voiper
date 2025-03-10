package app

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/megakuul/voiper/internal/config"
	"github.com/megakuul/voiper/internal/sip"
)

type App struct {
	ctx        context.Context
	basePath   string
	configLock sync.Mutex
	config     *config.Config
}

type AppOption func(*App)

func NewApp(opts ...AppOption) *App {
	app := &App{}

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
}

func (a *App) ListConfigs() (map[string]bool, error) {
	return config.ListConfigs(a.basePath)
}

func (a *App) AddConfig(cfg *config.Config, name, encryptionKey string) error {
	return config.WriteConfig(cfg, filepath.Join(a.basePath, name), encryptionKey)
}

func (a *App) RemoveConfig(path string, encrypted bool) error {
	return config.RemoveConfig(path, encrypted)
}

func (a *App) EnableConfig(path, decryptionKey string) error {
	cfg, err := config.LoadConfig(path, decryptionKey)
	if err != nil {
		return err
	}
	a.configLock.Lock()
	defer a.configLock.Unlock()
	a.config = cfg
	return nil
}

func (a *App) RegisterSIP() error {
	client := sip.NewClient(a.config)

	return client.Register()
}
