package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type DynamoSchema struct {
	TableName        string           `json:"table_name"`
	HashKey          string           `json:"hash_key"`
	RangeKey         string           `json:"range_key"`
	Attributes       []Attribute      `json:"attributes"`
	CommonAttributes []Attribute      `json:"common_attributes"`
	SecondaryIndexes []SecondaryIndex `json:"secondary_indexes"`
}

type Attribute struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SecondaryIndex struct {
	Name             string   `json:"name"`
	HashKey          string   `json:"hash_key"`
	RangeKey         string   `json:"range_key"`
	ProjectionType   string   `json:"projection_type"`
	NonKeyAttributes []string `json:"non_key_attributes,omitempty"`
}

const codeTemplate = `
// Code generated by dynamo_dictionary_table.go. DO NOT EDIT.

package {{.PackageName}}

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	TableName = "{{.TableName}}"
	{{range .SecondaryIndexes}}
	Index{{.Name}} = "{{.Name}}"
	{{- end}}
)

var (
	AttributeNames = []string{
		{{- range .AllAttributes}}
		"{{.Name}}",
		{{- end}}
	}

	IndexProjections = map[string][]string{
		{{- range .SecondaryIndexes}}
		Index{{.Name}}: {
			{{- if eq .ProjectionType "ALL"}}
			{{- range $.AllAttributes}}
			"{{.Name}}",
			{{- end}}
			{{- else}}
			"{{.HashKey}}", "{{.RangeKey}}",
			{{- range .NonKeyAttributes}}
			"{{.}}",
			{{- end}}
			{{- end}}
		},
		{{- end}}
	}
)

type DynamoSchema struct {
	TableName        string
	HashKey          string
	RangeKey         string
	Attributes       []Attribute
	CommonAttributes []Attribute
	SecondaryIndexes []SecondaryIndex
}

type Attribute struct {
	Name string
	Type string
}

type SecondaryIndex struct {
	Name             string
	HashKey          string
	RangeKey         string
	ProjectionType   string
	NonKeyAttributes []string
}

// SchemaItem represents the structure of the item in the table "{{.TableName}}"
type SchemaItem struct {
	{{range .AllAttributes}}
	{{.Name | ToCamelCase}} {{if eq .Type "S"}}string{{else if eq .Type "N"}}int{{else if eq .Type "B"}}bool{{else}}interface{}{{end}} ` + "`json:\"{{.Name}}\"`" + `
	{{end}}
}

var TableSchema = DynamoSchema{
	TableName: "{{.TableName}}",
	HashKey:   "{{.HashKey}}",
	RangeKey:  "{{.RangeKey}}",
	Attributes: []Attribute{
		{{- range .Attributes}}
		{Name: "{{.Name}}", Type: "{{.Type}}"},
		{{- end}}
	},
	CommonAttributes: []Attribute{
		{{- range .CommonAttributes}}
		{Name: "{{.Name}}", Type: "{{.Type}}"},
		{{- end}}
	},
	SecondaryIndexes: []SecondaryIndex{
		{{- range .SecondaryIndexes}}
		{
			Name:           "{{.Name}}",
			HashKey:        "{{.HashKey}}",
			RangeKey:       "{{.RangeKey}}",
			ProjectionType: "{{.ProjectionType}}",
			{{- if .NonKeyAttributes}}
			NonKeyAttributes: []string{
				{{- range .NonKeyAttributes}}
				"{{.}}",
				{{- end}}
			},
			{{- end}}
		},
		{{- end}}
	},
}

type QueryBuilder struct {
	IndexName       string
	KeyCondition    expression.KeyConditionBuilder
	FilterCondition expression.ConditionBuilder
	UsedKeys        map[string]bool
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		UsedKeys: make(map[string]bool),
	}
}

{{range .AllAttributes}}
func (qb *QueryBuilder) With{{.Name | ToCamelCase}}({{.Name | ToLowerCamelCase}} {{if eq .Type "N"}}int{{else if eq .Type "B"}}bool{{else}}string{{end}}) *QueryBuilder {
	{{- $attrName := .Name}}
	{{- range $.SecondaryIndexes}}
	{{- if eq .HashKey $attrName}}
	if qb.IndexName == "" {
		qb.IndexName = Index{{.Name}}
		qb.KeyCondition = expression.Key("{{$attrName}}").Equal(expression.Value({{$attrName | ToLowerCamelCase}}))
		qb.UsedKeys["{{$attrName}}"] = true
		return qb
	}
	{{- end}}
	{{- end}}
	if !qb.UsedKeys["{{$attrName}}"] {
		if qb.FilterCondition.IsSet() {
			qb.FilterCondition = qb.FilterCondition.And(expression.Name("{{$attrName}}").Equal(expression.Value({{$attrName | ToLowerCamelCase}})))
		} else {
			qb.FilterCondition = expression.Name("{{$attrName}}").Equal(expression.Value({{$attrName | ToLowerCamelCase}}))
		}
		qb.UsedKeys["{{$attrName}}"] = true
	}
	return qb
}
{{end}}

func (qb *QueryBuilder) Build() (string, expression.KeyConditionBuilder, expression.ConditionBuilder) {
	return qb.IndexName, qb.KeyCondition, qb.FilterCondition
}

// PutItem builds a map of AttributeValues for a DynamoDB PutItem operation
func PutItem(item SchemaItem) (map[string]types.AttributeValue, error) {
	attributeValues := make(map[string]types.AttributeValue)
	{{range .AllAttributes}}
	{{if eq .Type "N"}}
		attributeValues["{{.Name}}"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", item.{{.Name | ToCamelCase}})}
	{{else if eq .Type "B"}}
		attributeValues["{{.Name}}"] = &types.AttributeValueMemberBOOL{Value: item.{{.Name | ToCamelCase}}}
	{{else if eq .Type "S"}}
		if item.{{.Name | ToCamelCase}} != "" {
			attributeValues["{{.Name}}"] = &types.AttributeValueMemberS{Value: item.{{.Name | ToCamelCase}}}
		}
	{{end}}
	{{end}}
	return attributeValues, nil
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
`

func main() {
	rootDir, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("Error getting project root directory: %v\n", err)
		return
	}
	tmplDir := filepath.Join(rootDir, ".tmpl")

	files, err := os.ReadDir(tmplDir)
	if err != nil {
		fmt.Printf("Error reading template directory %s: %v\n", tmplDir, err)
		return
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			processSchemaFile(filepath.Join(tmplDir, file.Name()), rootDir)
		}
	}
}

func processSchemaFile(jsonPath, rootDir string) {
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading JSON file %s: %v\n", jsonPath, err)
		return
	}

	var schema DynamoSchema
	err = json.Unmarshal(jsonFile, &schema)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON from %s: %v\n", jsonPath, err)
		return
	}

	packageName := "gen_" + strings.ReplaceAll(schema.TableName, "-", "_")
	packageDir := filepath.Join(rootDir, "gen_"+strings.ReplaceAll(schema.TableName, "-", "_"))

	if err := os.MkdirAll(packageDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", packageDir, err)
		return
	}
	outputPath := filepath.Join(packageDir, packageName+".go")

	funcMap := template.FuncMap{
		"ToCamelCase":      toCamelCase,
		"ToLowerCamelCase": toLowerCamelCase,
	}
	allAttributes := append(schema.Attributes, schema.CommonAttributes...)

	schemaMap := map[string]interface{}{
		"PackageName":      packageName,
		"TableName":        schema.TableName,
		"HashKey":          schema.HashKey,
		"RangeKey":         schema.RangeKey,
		"Attributes":       schema.Attributes,
		"CommonAttributes": schema.CommonAttributes,
		"AllAttributes":    allAttributes,
		"SecondaryIndexes": schema.SecondaryIndexes,
	}

	tmpl, err := template.New("schema").Funcs(funcMap).Parse(codeTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file %s: %v\n", outputPath, err)
		return
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, schemaMap)
	if err != nil {
		fmt.Printf("Error executing template for %s: %v\n", outputPath, err)
		return
	}
	fmt.Printf("Generated code for %s successfully!\n", schema.TableName)
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func toLowerCamelCase(s string) string {
	s = toCamelCase(s)
	return strings.ToLower(s[:1]) + s[1:]
}
