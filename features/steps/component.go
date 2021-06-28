package steps

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/log.go/log"
	"github.com/maxcnunes/httpfake"
	"github.com/pkg/errors"
)

// Component contains all the information to create a component test
type Component struct {
	cfg *config.Config
	componenttest.ErrorFeature
	ctx         context.Context
	errorChan   chan error
	FakeAPI     *FakeAPI
	fakeRequest *httpfake.Request
	HTTPServer  *http.Server
	serviceList *service.ExternalServiceList
	signals     chan os.Signal
	svc         *service.Service
	svcErrors   chan error
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (*Component, error) {
	c := &Component{
		ctx:        context.Background(),
		errorChan:  make(chan error),
		HTTPServer: &http.Server{},
	}

	c.FakeAPI = NewFakeAPI(c)

	if err := c.initialise(); err != nil {
		return nil, errors.Wrap(err, "failed to initialise component")
	}

	c.run()

	return c, nil
}

func (c *Component) initialise() (err error) {
	c.signals = make(chan os.Signal, 1)
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)

	c.svcErrors = make(chan error, 1)

	c.cfg, err = config.Get()
	if err != nil {
		return err
	}

	c.cfg.APIRouterURL = c.FakeAPI.fakeHTTP.ResolveURL("")

	initFunctions := &mocks.InitialiserMock{
		DoGetHTTPServerFunc:   c.getHTTPServer,
		DoGetHealthCheckFunc:  getHealthCheck,
		DoGetHealthClientFunc: getHealthClient,
	}

	c.serviceList = service.NewServiceList(initFunctions)

	return nil
}

func (c *Component) getHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func getHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func getHealthClient(name string, url string) *health.Client {
	return &health.Client{}
}

func (c *Component) run() {
	go func() error {
		c.svc = service.New()

		if err := c.svc.Init(c.ctx, c.cfg, c.serviceList); err != nil {
			return errors.Wrap(err, "failed to initialise service")
		}

		c.svc.Start(c.ctx, c.svcErrors)

		// blocks until an os interrupt or a fatal error occurs
		select {
		case err := <-c.errorChan:
			log.Event(c.ctx, "service error received", log.ERROR, log.Error(err))
		case sig := <-c.signals:
			log.Event(c.ctx, "os signal received", log.Data{"signal": sig}, log.INFO)
		}

		return c.svc.Close(c.ctx)
	}()
}
