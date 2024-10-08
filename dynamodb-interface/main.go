package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
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
// Code generated by code generator. DO NOT EDIT.

package {{.PackageName}}

import (
    "fmt"

    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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
        "{{.Name}}": {
            {{- if eq .ProjectionType "ALL"}}
            {{- range $.AllAttributes}}
            "{{.Name}}",
            {{- end}}
            {{- else}}
            "{{.HashKey}}", {{if .RangeKey}}"{{.RangeKey}}",{{end}}
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

// SchemaItem represents an item in "{{.TableName}}"
type SchemaItem struct {
    {{range .AllAttributes}}
    {{SafeName .Name | ToCamelCase}} {{TypeGo .Type}} ` + "`dynamodbav:\"{{.Name}}\"`" + `
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
    IndexName        string
    KeyConditions    []expression.KeyConditionBuilder
    FilterConditions []expression.ConditionBuilder
    UsedKeys         map[string]bool
    Attributes       map[string]interface{}
}

func NewQueryBuilder() *QueryBuilder {
    return &QueryBuilder{
        UsedKeys:   make(map[string]bool),
        Attributes: make(map[string]interface{}),
    }
}

{{range .AllAttributes}}
func (qb *QueryBuilder) With{{SafeName .Name | ToCamelCase}}({{SafeName .Name | ToLowerCamelCase}} {{TypeGo .Type}}) *QueryBuilder {
    attrName := "{{.Name}}"
    qb.Attributes[attrName] = {{SafeName .Name | ToLowerCamelCase}}
    qb.UsedKeys[attrName] = true
    return qb
}
{{end}}

func (qb *QueryBuilder) Build() (string, expression.KeyConditionBuilder, *expression.ConditionBuilder, error) {
    var index *SecondaryIndex
    var keyCond expression.KeyConditionBuilder
    var filterCond *expression.ConditionBuilder

    // Try to find an index that matches the provided keys
    for _, idx := range TableSchema.SecondaryIndexes {
        if qb.UsedKeys[idx.HashKey] {
            index = &idx
            // Build KeyCondition
            keyCond = expression.Key(idx.HashKey).Equal(expression.Value(qb.Attributes[idx.HashKey]))
            if idx.RangeKey != "" && qb.UsedKeys[idx.RangeKey] {
                keyCond = keyCond.And(expression.Key(idx.RangeKey).Equal(expression.Value(qb.Attributes[idx.RangeKey])))
            }
            break
        }
    }

    // If no secondary index matches, try the primary key
    if index == nil && qb.UsedKeys[TableSchema.HashKey] {
        indexName := ""
        keyCond = expression.Key(TableSchema.HashKey).Equal(expression.Value(qb.Attributes[TableSchema.HashKey]))
        if TableSchema.RangeKey != "" && qb.UsedKeys[TableSchema.RangeKey] {
            keyCond = keyCond.And(expression.Key(TableSchema.RangeKey).Equal(expression.Value(qb.Attributes[TableSchema.RangeKey])))
        }
        // Build FilterCondition for remaining attributes
        for attrName, value := range qb.Attributes {
            if attrName != TableSchema.HashKey && attrName != TableSchema.RangeKey {
                cond := expression.Name(attrName).Equal(expression.Value(value))
                qb.FilterConditions = append(qb.FilterConditions, cond)
            }
        }
        // Combine filter conditions if any
        if len(qb.FilterConditions) > 0 {
            combinedFilter := qb.FilterConditions[0]
            for _, cond := range qb.FilterConditions[1:] {
                combinedFilter = combinedFilter.And(cond)
            }
            filterCond = &combinedFilter
        }
        return indexName, keyCond, filterCond, nil
    }

    if index == nil {
        return "", expression.KeyConditionBuilder{}, nil, fmt.Errorf("no suitable index found for the provided keys")
    }

    // Build FilterCondition for the remaining attributes
    for attrName, value := range qb.Attributes {
        if attrName != index.HashKey && attrName != index.RangeKey {
            cond := expression.Name(attrName).Equal(expression.Value(value))
            qb.FilterConditions = append(qb.FilterConditions, cond)
        }
    }

    // Combine filter conditions if any
    if len(qb.FilterConditions) > 0 {
        combinedFilter := qb.FilterConditions[0]
        for _, cond := range qb.FilterConditions[1:] {
            combinedFilter = combinedFilter.And(cond)
        }
        filterCond = &combinedFilter
    }

    return index.Name, keyCond, filterCond, nil
}

// PutItem creates an AttributeValues map for PutItem in DynamoDB
func PutItem(item SchemaItem) (map[string]types.AttributeValue, error) {
    attributeValues, err := attributevalue.MarshalMap(item)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal item: %v", err)
    }
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
		fmt.Printf("Cannot find root directory: %v\n", err)
		return
	}
	tmplDir := filepath.Join(rootDir, ".tmpl")

	files, err := os.ReadDir(tmplDir)
	if err != nil {
		fmt.Printf("Read template directory failed %s: %v\n", tmplDir, err)
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
		fmt.Printf("Failed to read json %s: %v\n", jsonPath, err)
		return
	}

	var schema DynamoSchema
	err = json.Unmarshal(jsonFile, &schema)
	if err != nil {
		fmt.Printf("Failed to parse json %s: %v\n", jsonPath, err)
		return
	}

	packageName := strings.ReplaceAll(schema.TableName, "-", "")
	packageDir := filepath.Join(rootDir, "gen", packageName)

	if err := os.MkdirAll(packageDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", packageDir, err)
		return
	}
	outputPath := filepath.Join(packageDir, packageName+".go")

	funcMap := template.FuncMap{
		"ToCamelCase":      toCamelCase,
		"ToLowerCamelCase": toLowerCamelCase,
		"SafeName":         safeName,
		"TypeGo":           typeGo,
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
		fmt.Printf("Failed to parse template: %v\n", err)
		return
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Failed to create output file %s: %v\n", outputPath, err)
		return
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, schemaMap)
	if err != nil {
		fmt.Printf("Failed to execute template for %s: %v\n", outputPath, err)
		return
	}
	fmt.Printf("Successfully generated %s!\n", schema.TableName)
}

func toCamelCase(s string) string {
	var result string
	capitalizeNext := true
	for _, r := range s {
		if r == '_' || r == '-' {
			capitalizeNext = true
		} else if capitalizeNext {
			result += string(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result += string(r)
		}
	}
	return result
}

func toLowerCamelCase(s string) string {
	if s == "" {
		return ""
	}
	s = toCamelCase(s)
	return strings.ToLower(s[:1]) + s[1:]
}

var reservedWords = map[string]bool{
	// List of Go reserved words
	"break":       true,
	"default":     true,
	"func":        true,
	"interface":   true,
	"select":      true,
	"case":        true,
	"defer":       true,
	"go":          true,
	"map":         true,
	"struct":      true,
	"chan":        true,
	"else":        true,
	"goto":        true,
	"package":     true,
	"switch":      true,
	"const":       true,
	"fallthrough": true,
	"if":          true,
	"range":       true,
	"type":        true,
	"continue":    true,
	"for":         true,
	"import":      true,
	"return":      true,
	"var":         true,
}

func safeName(s string) string {
	if reservedWords[s] {
		return s + "_"
	}
	return s
}

func typeGo(dynamoType string) string {
	switch dynamoType {
	case "S":
		return "string"
	case "N":
		return "int"
	case "B":
		return "bool"
	default:
		return "interface{}"
	}
}
