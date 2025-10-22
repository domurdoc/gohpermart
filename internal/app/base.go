package app

import (
	"github.com/domurdoc/gophermart/internal/config"
	"github.com/domurdoc/gophermart/internal/utils"
)

type App struct {
	Options      *config.Options
	Repositories *Repositories
	Services     *Services

	closer *utils.Closer
}

func New() (*App, error) {
	a := &App{Options: config.New(), closer: utils.NewCloser()}

	if err := a.initRepositories(); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.initServices(); err != nil {
		a.Close()
		return nil, err
	}
	return a, nil
}

func (a *App) Close() error {
	return a.closer.Close()
}

func (a *App) initRepositories() error {
	repos, err := NewRepositories(&a.Options.Repositories)
	if err != nil {
		return err
	}
	a.Repositories = repos
	a.closer.Register(a.Repositories.Close)
	return nil
}

func (a *App) initServices() error {
	services, err := NewServices(&a.Options.Services, a.Repositories)
	if err != nil {
		return err
	}
	a.Services = services
	a.closer.Register(a.Services.Close)
	return nil
}
