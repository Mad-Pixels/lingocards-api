package main

import (
	"context"
	"encoding/json"
	"github.com/Mad-Pixels/lingocards-api/pkg/tools"
	"net/http"

	"github.com/Mad-Pixels/lingocards-api/pkg/api"
	"github.com/Mad-Pixels/lingocards-api/pkg/serializer"

	"github.com/Mad-Pixels/lingocards-api/dynamodb-interface/gen/lingocardsdictionary"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/rs/zerolog"
)

type handleDataDeleteRequest struct {
	Name   string `json:"name" validate:"required,min=4,max=32"`
	Author string `json:"author" validate:"required"`
}

type handleDataDeleteResponse struct {
	Status string `json:"status"`
}

func handleDataDelete(ctx context.Context, _ zerolog.Logger, raw json.RawMessage) (any, *api.HandleError) {
	var req handleDataDeleteRequest
	if err := serializer.UnmarshalJSON(raw, &req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}
	if err := validate.Struct(&req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}

	var (
		id  = tools.EncodeBaseID(req.Name, req.Author)
		key = map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: id},
			"name": &types.AttributeValueMemberS{Value: req.Name},
		}
	)
	result, err := dbDynamo.Get(ctx, lingocardsdictionary.TableSchema.TableName, key)
	if err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if len(result.Item) == 0 {
		return nil, &api.HandleError{Status: http.StatusNotFound, Err: err}
	}

	var item lingocardsdictionary.SchemaItem
	if err = attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if err = s3Bucket.Delete(ctx, item.Filename, serviceDictionaryBucket); err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if err = dbDynamo.Delete(ctx, lingocardsdictionary.TableSchema.TableName, key); err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	return handleDataDeleteResponse{Status: "OK"}, nil
}
