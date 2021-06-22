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
	"github.com/maxcnunes/httpfake"
)

// Component contains all the information to create a component test
type Component struct {
	componenttest.ErrorFeature
	ctx         context.Context
	errorChan   chan error
	FakeAPI     *FakeAPI
	fakeRequest *httpfake.Request
	HTTPServer  *http.Server
	svc         *service.Service
}

// NewSearchControllerComponent creates a search controller component
func NewSearchControllerComponent() (*Component, error) {
	c := &Component{
		HTTPServer: &http.Server{},
		errorChan:  make(chan error),
		ctx:        context.Background(),
	}

	c.FakeAPI = NewFakeAPI(c)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	cfg.APIRouterURL = c.FakeAPI.fakeHTTP.ResolveURL("")

	initFunctions := &initialiser.InitialiserMock{
		DoGetHTTPServerFunc:   c.DoGetHTTPServer,
		DoGetHealthCheckFunc:  DoGetHealthcheckOk,
		DoGetHealthClientFunc: DoGetHealthClient,
	}

	serviceList := service.NewServiceList(initFunctions)

	c.runApplication(cfg, serviceList, signals)

	return c, nil
}
