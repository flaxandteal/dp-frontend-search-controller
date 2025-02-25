package service_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/renderer"
	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-frontend-search-controller/service"
	"github.com/ONSdigital/dp-frontend-search-controller/service/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx = context.Background()

	errAddCheckFail = errors.New("Error(s) registering checkers for healthcheck")
	errHealthCheck  = errors.New("healthCheck error")
	errServer       = errors.New("HTTP Server error")

	// Health Check Mock
	hcMock = &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}
	hcMockAddFail = &mocks.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddCheckFail },
		StartFunc:    func(ctx context.Context) {},
	}
	funcDoGetHealthCheckOK = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return hcMock, nil
	}
	funcDoGetHealthCheckFail = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return nil, errHealthCheck
	}
	funcDoGetHealthAddCheckerFail = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return hcMockAddFail, nil
	}

	// Server Mock
	serverWg   = &sync.WaitGroup{}
	serverMock = &mocks.HTTPServerMock{
		ListenAndServeFunc: func() error {
			serverWg.Done()
			return nil
		},
	}
	failingServerMock = &mocks.HTTPServerMock{
		ListenAndServeFunc: func() error {
			serverWg.Done()
			return errServer
		},
	}
	funcDoGetHTTPServerOK = func(bindAddr string, router http.Handler) service.HTTPServer {
		return serverMock
	}
	funcDoGetHTTPServerFail = func(bindAddr string, router http.Handler) service.HTTPServer {
		return failingServerMock
	}

	// Health Client Mock
	funcDoGetHealthClient = func(name string, url string) *health.Client {
		return &health.Client{
			URL:    url,
			Name:   name,
			Client: service.NewMockHTTPClient(&http.Response{}, nil),
		}
	}

	// Renderer Client Mock
	funcDoGetRendererClientOK = func(rendererURL string) *renderer.Renderer {
		return &renderer.Renderer{
			HcCli: &health.Client{
				URL:    rendererURL,
				Name:   "renderer",
				Client: service.NewMockHTTPClient(&http.Response{}, nil),
			},
		}
	}
)

func TestNew(t *testing.T) {
	Convey("New returns a new uninitialised service", t, func() {
		So(service.New(), ShouldResemble, &service.Service{})
	})
}

func TestInitSuccess(t *testing.T) {
	Convey("Given all dependencies are successfully initialised", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckOK,
			DoGetHTTPServerFunc:   funcDoGetHTTPServerOK,
		}
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svc := &service.Service{}

			Convey("When Init is called", func() {
				err := svc.Init(ctx, cfg, mockServiceList)

				Convey("Then service is initialised successfully", func() {
					So(svc.Config, ShouldResemble, cfg)
					So(svc.HealthCheck, ShouldResemble, hcMock)
					So(svc.Server, ShouldResemble, serverMock)
					So(svc.ServiceList, ShouldResemble, mockServiceList)

					Convey("And returns no errors", func() {
						So(err, ShouldBeNil)

						Convey("And the checkers are registered and the healthcheck", func() {
							So(mockServiceList.HealthCheck, ShouldBeTrue)
							So(len(hcMock.AddCheckCalls()), ShouldEqual, 1)
							So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "API router")
							So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
							So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:25000")
						})
					})
				})
			})
		})
	})
}

func TestInitFailure(t *testing.T) {
	Convey("Given failure to create healthcheck", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckFail,
		}
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svc := &service.Service{}

			Convey("When Init is called", func() {
				err := svc.Init(ctx, cfg, mockServiceList)

				Convey("Then service initialisation fails", func() {
					So(svc.Config, ShouldResemble, cfg)
					So(svc.ServiceList, ShouldResemble, mockServiceList)
					So(svc.ServiceList.HealthCheck, ShouldBeFalse)

					// Healthcheck and Server not initialised
					So(svc.HealthCheck, ShouldBeNil)
					So(svc.Server, ShouldBeNil)

					Convey("And returns error", func() {
						So(err, ShouldNotBeNil)
						So(err, ShouldResemble, errHealthCheck)
					})
				})
			})
		})
	})

	Convey("Given that Checkers cannot be registered", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthAddCheckerFail,
		}
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svc := &service.Service{}

			Convey("When Init is called", func() {
				err := svc.Init(ctx, cfg, mockServiceList)

				Convey("Then service initialisation fails", func() {
					So(svc.Config, ShouldResemble, cfg)
					So(svc.ServiceList, ShouldResemble, mockServiceList)
					So(svc.ServiceList.HealthCheck, ShouldBeTrue)
					So(svc.HealthCheck, ShouldResemble, hcMockAddFail)

					// Server not initialised
					So(svc.Server, ShouldBeNil)

					Convey("And returns error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldResemble, errAddCheckFail.Error())

						Convey("And all checks try to register", func() {
							So(mockServiceList.HealthCheck, ShouldBeTrue)
							So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 1)
							So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "API router")
						})
					})
				})
			})
		})
	})
}

func TestStart(t *testing.T) {
	Convey("Given a correctly initialised Service with mocked dependencies", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckOK,
			DoGetHTTPServerFunc:   funcDoGetHTTPServerOK,
		}
		serverWg.Add(1)
		mockServiceList := service.NewServiceList(initMock)

		svcErrors := make(chan error, 1)

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		svc := &service.Service{
			Config:      cfg,
			HealthCheck: hcMock,
			Server:      serverMock,
			ServiceList: mockServiceList,
		}

		Convey("When service starts", func() {
			svc.Run(ctx, svcErrors)

			Convey("Then healthcheck is started and HTTP server starts listening", func() {
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given that HTTP Server fails", t, func() {
		initMock := &mocks.InitialiserMock{
			DoGetHealthClientFunc: funcDoGetHealthClient,
			DoGetHealthCheckFunc:  funcDoGetHealthCheckOK,
			DoGetHTTPServerFunc:   funcDoGetHTTPServerFail,
		}
		serverWg.Add(1)
		mockServiceList := service.NewServiceList(initMock)

		Convey("and valid config and service error channel are provided", func() {
			service.BuildTime = "TestBuildTime"
			service.GitCommit = "TestGitCommit"
			service.Version = "TestVersion"

			svcErrors := make(chan error, 1)

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			svc := &service.Service{
				Config:      cfg,
				HealthCheck: hcMock,
				Server:      failingServerMock,
				ServiceList: mockServiceList,
			}

			Convey("When service starts", func() {
				svc.Run(ctx, svcErrors)

				Convey("Then service start fails and returns an error in the error channel", func() {
					sErr := <-svcErrors
					So(sErr.Error(), ShouldResemble, errServer.Error())
					So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}

func TestCloseSuccess(t *testing.T) {
	Convey("Given a correctly initialised service", t, func() {

		ctx := context.Background()

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcCloseMock := &mocks.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverCloseMock := &mocks.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		serviceList := service.NewServiceList(nil)
		serviceList.HealthCheck = true
		svc := service.Service{
			Config:      cfg,
			HealthCheck: hcCloseMock,
			Server:      serverCloseMock,
			ServiceList: serviceList,
		}

		Convey("When closing service", func() {
			err = svc.Close(ctx)

			Convey("Then it results in all the dependencies being closed in the expected order", func() {
				So(err, ShouldBeNil)
				So(len(hcCloseMock.StopCalls()), ShouldEqual, 1)
				So(len(serverCloseMock.ShutdownCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestCloseFailure(t *testing.T) {
	Convey("Given if service fails to stop", t, func() {
		failingServerCloseMock := &mocks.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				return errors.New("Failed to stop http server")
			},
		}

		Convey("And given a correctly initialised service", func() {
			ctx := context.Background()

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			serviceList := service.NewServiceList(nil)
			serviceList.HealthCheck = true
			svc := service.Service{
				Config:      cfg,
				HealthCheck: hcMock,
				Server:      failingServerCloseMock,
				ServiceList: serviceList,
			}

			Convey("When closing the service", func() {
				err = svc.Close(ctx)

				Convey("Then Close operation tries to close all dependencies", func() {
					So(len(hcMock.StopCalls()), ShouldEqual, 1)
					So(len(failingServerCloseMock.ShutdownCalls()), ShouldEqual, 1)

					Convey("And returns an error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
					})
				})
			})
		})
	})

	Convey("Given that a dependency takes more time to close than the graceful shutdown timeout", t, func() {
		hcStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcShutdownCloseMock := &mocks.HealthCheckerMock{
			StopFunc: func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverShutdownCloseMock := &mocks.HTTPServerMock{
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server was stopped before healthcheck")
				}
				return nil
			},
		}

		serverShutdownCloseMock.ShutdownFunc = func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			return nil
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.GracefulShutdownTimeout = 1 * time.Millisecond

		Convey("And given a correctly initialised service", func() {
			ctx := context.Background()

			serviceList := service.NewServiceList(nil)
			serviceList.HealthCheck = true
			svc := service.Service{
				Config:      cfg,
				HealthCheck: hcShutdownCloseMock,
				Server:      serverShutdownCloseMock,
				ServiceList: serviceList,
			}

			Convey("When closing the service", func() {
				err = svc.Close(ctx)

				Convey("Then closing the service fails with context.DeadlineExceeded error and no further dependencies are attempted to close", func() {
					So(err, ShouldResemble, context.DeadlineExceeded)
					So(len(hcShutdownCloseMock.StopCalls()), ShouldEqual, 1)
					So(len(serverShutdownCloseMock.ShutdownCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
