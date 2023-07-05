package main

import (
	"context"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/vs49688/servicebase"
	"github.com/vs49688/servicebase/cmd/sample/pb"
)

type sampleService struct {
	pb.UnimplementedTeapotServer

	logger *log.Logger
}

func (d *sampleService) Close(_ context.Context) error {
	return nil
}

func (d *sampleService) GetHealth(_ context.Context) (*servicebase.GetHealthResponse, error) {
	return &servicebase.GetHealthResponse{
		Status: servicebase.HealthStatusHealthy,
	}, nil
}

func (d *sampleService) AmIATeapot(ctx context.Context, _ *pb.AmIATeapotRequest) (*pb.AmIATeapotResponse, error) {
	// grpcurl -plaintext localhost:50051 sample.Teapot.AmIATeapot
	d.logger.WithContext(ctx).Info("im a teapot")
	return &pb.AmIATeapotResponse{Answer: true}, nil
}

func (d *sampleService) httpTeapot(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTeapot)
	_, _ = w.Write([]byte("im a teapot"))
}

func makeService(_ *servicebase.ServiceConfig) servicebase.ServiceFactory {
	return func(ctx context.Context, params servicebase.ServiceParameters) (servicebase.Service, error) {
		svc := &sampleService{
			logger: params.Logger,
		}
		params.ApplicationRouter.HandleFunc("/teapot", svc.httpTeapot)
		pb.RegisterTeapotServer(params.GRPCRegistrar, svc)
		return svc, nil
	}
}

func main() {
	cfg := servicebase.DefaultServiceConfig()
	cfg.GRPC.EnableReflection = true

	app := &cli.App{
		Name:                   "servicebase-sample",
		Usage:                  os.Args[0],
		Description:            "Sample Application for servicebase",
		Flags:                  cfg.Flags(),
		UseShortOptionHandling: true,
		Action: func(context *cli.Context) error {
			return servicebase.RunService(context.Context, cfg, makeService(&cfg))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
