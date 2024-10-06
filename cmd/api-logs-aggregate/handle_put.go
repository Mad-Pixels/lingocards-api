package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Mad-Pixels/lingocards-api/dynamodb-interface/gen/lingocardslogs"
	"github.com/Mad-Pixels/lingocards-api/pkg/tools"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"net/http"
	"strconv"

	"github.com/Mad-Pixels/lingocards-api/pkg/api"
	"github.com/Mad-Pixels/lingocards-api/pkg/serializer"
	"github.com/rs/zerolog"
)

type handleDataLogsRequest struct {
	Timestamp      int    `json:"timestamp" validate:"required"`
	OsVersion      string `json:"os_version" validate:"required"`
	Device         string `json:"device" validate:"required"`
	ErrorType      string `json:"error_type" validate:"required"`
	ErrorMessage   string `json:"error_message" validate:"required"`
	AppVersion     string `json:"app_version" validate:"required"`
	AdditionalInfo string `json:"additional_info"`
}

type handleDataLogsResponse struct {
	Status string `json:"status"`
}

func handlePut(ctx context.Context, _ zerolog.Logger, raw json.RawMessage) (any, *api.HandleError) {
	var req handleDataLogsRequest
	if err := serializer.UnmarshalJSON(raw, &req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}
	if err := validate.Struct(&req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}
	var (
		id  = tools.EncodeBaseID(req.ErrorType, req.Device, req.OsVersion)
		key = map[string]types.AttributeValue{
			"id":         &types.AttributeValueMemberS{Value: id},
			"error_type": &types.AttributeValueMemberS{Value: req.ErrorType},
		}
	)

	exist, err := dbDynamo.Get(ctx, lingocardslogs.TableSchema.TableName, key)
	if err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if exist != nil && len(exist.Item) > 0 {
		count := 0
		if val, ok := exist.Item["count"]; ok {
			count, _ = strconv.Atoi(val.(*types.AttributeValueMemberN).Value)
		}
		count++
		updates := map[string]types.AttributeValue{
			"count":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", count)},
			"last_timestamp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", req.Timestamp)},
		}

		condition := expression.AttributeExists(expression.Name("count"))
		if err = dbDynamo.Update(ctx, lingocardslogs.TableSchema.TableName, key, updates, condition); err != nil {
			return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
		}
		return handleDataLogsResponse{Status: "OK"}, nil
	}

	schemaItem := lingocardslogs.SchemaItem{
		LastTimestamp:  req.Timestamp,
		FirstTimestamp: req.Timestamp,
		OsVersion:      req.OsVersion,
		Device:         req.Device,
		ErrorType:      req.ErrorType,
		ErrorMessage:   req.ErrorMessage,
		AppVersion:     req.AppVersion,
		AdditionalInfo: req.AdditionalInfo,
		Id:             id,
		Count:          1,
	}
	item, err := lingocardslogs.PutItem(schemaItem)
	if err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}

	condition := expression.AttributeNotExists(expression.Name("id"))
	if err = dbDynamo.Put(ctx, lingocardslogs.TableSchema.TableName, item, condition); err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	return handleDataLogsResponse{Status: "OK"}, nil
}
