package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/Mad-Pixels/lingocards-api/pkg/api"
	"github.com/Mad-Pixels/lingocards-api/pkg/cloud"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-playground/validator/v10"
)

var (
	awsRegion = os.Getenv("AWS_REGION")

	validate *validator.Validate
	dbDynamo *cloud.Dynamo
)

func init() {
	debug.SetGCPercent(500)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		panic("unable to load AWS SDK config: " + err.Error())
	}
	dbDynamo = cloud.NewDynamo(cfg)
	validate = validator.New()
}

func main() {
	lambda.Start(
		api.NewLambda(
			api.Config{},
			map[string]api.HandleFunc{
				"put": handlePut,
			},
		).Handle,
	)
}
