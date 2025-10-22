package app

import (
	"errors"

	"github.com/domurdoc/gophermart/internal/config"
)

type App struct {
	Options      *config.Options
	Repositories *Repositories
	Services     *Services
}

func New() (*App, error) {
	a := &App{Options: config.New()}

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
	var errs []error

	if a.Repositories != nil {
		errs = append(errs, a.Repositories.Close())
	}
	if a.Services != nil {
		errs = append(errs, a.Services.Close())
	}
	return errors.Join(errs...)
}

func (a *App) initRepositories() error {
	repos, err := NewRepositories(&a.Options.Repositories)
	if err != nil {
		return err
	}
	a.Repositories = repos
	return nil
}

func (a *App) initServices() error {
	services, err := NewServices(&a.Options.Services, a.Repositories)
	if err != nil {
		return err
	}
	a.Services = services
	return nil
}
