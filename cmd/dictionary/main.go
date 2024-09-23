package main

import (
	"context"
	"os"

	"github.com/Mad-Pixels/lingocards-api/internal/lambda"
	"github.com/Mad-Pixels/lingocards-api/pkg/cloud"
	aws_lambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	validator "github.com/go-playground/validator/v10"
)

var (
	// service vars.
	serviceDictionaryBucket = os.Getenv("SERVICE_DICTIONARY_BUCKET")
	serviceProcessingBucket = os.Getenv("SERVICE_PROCESSING_BUCKET")

	// system vars.
	awsRegion = os.Getenv("AWS_REGION")
	token     = os.Getenv("AUTH_TOKEN")
	validate  *validator.Validate
	s3Bucket  *cloud.Bucket
	dbDynamo  *cloud.Dynamo
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		panic("unable to load AWS SDK config: " + err.Error())
	}
	s3Bucket = cloud.NewBucket(cfg)
	dbDynamo = cloud.NewDynamo(cfg)
	validate = validator.New()
}

func main() {
	aws_lambda.Start(
		lambda.NewLambda(
			lambda.Config{Token: token},
			map[string]lambda.HandleFunc{
				"file_presign": handleFilePresign,
				"data_delete":  handleDataDelete,
				"data_query":   handleDataQuery,
				"data_put":     handleDataPut,
			},
		).Handle,
	)
}
