package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/Mad-Pixels/lingocards-api/pkg/api"
	"github.com/go-playground/validator/v10"

	"github.com/Mad-Pixels/lingocards-api/pkg/cloud"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	serviceDictionaryBucket = os.Getenv("SERVICE_DICTIONARY_BUCKET")
	awsRegion               = os.Getenv("AWS_REGION")

	validate *validator.Validate
	s3Bucket *cloud.Bucket
	dbDynamo *cloud.Dynamo
)

func init() {
	debug.SetGCPercent(500)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		panic("unable to load AWS SDK config: " + err.Error())
	}
	s3Bucket = cloud.NewBucket(cfg)
	dbDynamo = cloud.NewDynamo(cfg)
	validate = validator.New()
}

func main() {
	lambda.Start(
		api.NewLambda(
			api.Config{},
			map[string]api.HandleFunc{
				"query":        handleDataQuery,
				"download_url": handleDownloadUrl,
			},
		).Handle,
	)
}
