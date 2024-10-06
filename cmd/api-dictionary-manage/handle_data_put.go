package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Mad-Pixels/lingocards-api/pkg/tools"
	"net/http"

	"github.com/Mad-Pixels/lingocards-api/pkg/api"
	"github.com/Mad-Pixels/lingocards-api/pkg/serializer"
	"github.com/go-playground/validator/v10"

	"github.com/Mad-Pixels/lingocards-api/dynamodb-interface/gen/lingocardsdictionary"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog"
)

const (
	defaultCode = "0"
)

type handleDataPutRequest struct {
	Description  string `json:"description" validate:"required"`
	Filename     string `json:"filename" validate:"required"`
	Name         string `json:"name" validate:"required,min=4,max=32"`
	Author       string `json:"author" validate:"required"`
	Code         string `json:"code,omitempty" validate:"validate_code"`
	CategoryMain string `json:"category_main" validate:"required"`
	CategorySub  string `json:"category_sub" validate:"required"`
	Public       bool   `json:"is_public" validate:"required"`
}

// custom validator tag.
func validateCode(fl validator.FieldLevel) bool {
	r, ok := fl.Parent().Interface().(handleDataPutRequest)
	if !ok {
		return false
	}
	return (r.Public && r.Code == "") || (!r.Public && r.Code != "")
}

type handleDataPutResponse struct {
	Status string `json:"status"`
}

func handleDataPut(ctx context.Context, _ zerolog.Logger, raw json.RawMessage) (any, *api.HandleError) {
	var req handleDataPutRequest
	if err := serializer.UnmarshalJSON(raw, &req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}
	if err := validate.RegisterValidation("validate_code", validateCode); err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if err := validate.Struct(&req); err != nil {
		return nil, &api.HandleError{Status: http.StatusBadRequest, Err: err}
	}
	id := tools.EncodeBaseID(req.Name, req.Author)

	schemaItem := lingocardsdictionary.SchemaItem{
		Id:           id,
		Name:         req.Name,
		Author:       req.Author,
		Filename:     req.Filename,
		CategoryMain: req.CategoryMain,
		CategorySub:  req.CategorySub,
		Description:  req.Description,
		IsPublic:     lingocardsdictionary.BoolToInt(req.Public),

		Code: func(val, defaultVal string) string {
			if val == "" {
				return defaultVal
			}
			return val
		}(req.Code, defaultCode),
	}

	item, err := lingocardsdictionary.PutItem(schemaItem)
	if err != nil {
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	if err = dbDynamo.Put(ctx, lingocardsdictionary.TableSchema.TableName, item, expression.AttributeNotExists(expression.Name("id"))); err != nil {
		var cfe *types.ConditionalCheckFailedException

		if errors.As(err, &cfe) {
			return nil, &api.HandleError{Status: http.StatusConflict, Err: errors.New("dictionary with id: '" + id + "' already exists")}
		}
		return nil, &api.HandleError{Status: http.StatusInternalServerError, Err: err}
	}
	return handleDataPutResponse{Status: "OK"}, nil
}
