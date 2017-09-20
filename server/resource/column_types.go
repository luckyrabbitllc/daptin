package resource

import (
	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"
	"time"
	"math/rand"
	"fmt"
	validator2 "gopkg.in/go-playground/validator.v9"
)

type Faker interface {
	Fake() string
}

type ColumnType struct {
	BlueprintType string
	Name          string
	Validations   []string
	Conformations []string
	ReclineType   string
	DataTypes     []string
}

func randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func (ct ColumnType) Fake() interface{} {

	switch ct.Name {
	case "id":
		return uuid.NewV4().String()
	case "alias":
		return uuid.NewV4().String()
	case "date":
		return randate().Format("2006-01-02")
	case "time":
		return randate().Format("15:04:05")
	case "day":
		return fake.Day()
	case "month":
		return fake.Month()
	case "year":
		return fake.Year(1990, 2018)
	case "minute":
		return rand.Intn(60)
	case "hour":
		return rand.Intn(24)
	case "datetime":
		return randate().Format(time.RFC3339)
	case "email":
		return fake.EmailAddress()
	case "name":
		return fake.FullName()
	case "json":
		return "{}"
	case "password":
		return ""
	case "value":
		return rand.Intn(1000)
	case "truefalse":
		return rand.Intn(3) == 1
	case "timestamp":
		return randate().Unix()
	case "location.latitude":
		return fake.Latitude()
	case "location":
		return fmt.Sprintf("[%v, %v]", fake.Latitude(), fake.Longitude())
	case "location.longitude":
		return fake.Longitude()
	case "location.altitude":
		return rand.Intn(10000)
	case "color":
		return fake.HexColor()
	case "rating.10":
		return rand.Intn(11)
	case "measurement":
		return rand.Intn(5000)
	case "label":
		return fake.ProductName()
	case "content":
		return fake.Sentences()
	case "file":
		return ""
	case "url":
		return "https://places.com/"
	default:
		return ""
	}
}

/**
"string"
"number"
"integer"
"date"
"time"
"date-time"
"boolean"
"binary"
"geo_point"
 */

var ColumnTypes = []ColumnType{
	{
		Name:          "id",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{},
		DataTypes:     []string{"varchar(20)", "varchar(10)"},
	},
	{
		Name:          "alias",
		BlueprintType: "string",
		ReclineType:   "string",
		DataTypes:     []string{"varchar(100)", "varchar(20)", "varchar(10)"},
	},
	{
		Name:          "date",
		BlueprintType: "string",
		ReclineType:   "date",
		DataTypes:     []string{"timestamp"},
	},
	{
		Name:          "time",
		BlueprintType: "string",
		ReclineType:   "time",
		DataTypes:     []string{"timestamp"},
	},
	{
		Name:          "day",
		BlueprintType: "string",
		ReclineType:   "string",
		DataTypes:     []string{"varchar(10)"},
	},
	{
		Name:          "month",
		BlueprintType: "number",
		ReclineType:   "string",
		Validations:   []string{"min=1,max=12"},
		DataTypes:     []string{"int(4)"},
	},
	{
		Name:          "year",
		BlueprintType: "number",
		ReclineType:   "string",
		Validations:   []string{"min=1900,max=2100"},
		DataTypes:     []string{"int(4)"},
	},
	{
		Name:          "minute",
		BlueprintType: "number",
		Validations:   []string{"min=0,max=59"},
		DataTypes:     []string{"int(4)"},
	},
	{
		Name:          "hour",
		BlueprintType: "number",
		ReclineType:   "string",
		DataTypes:     []string{"int(4)"},
	},
	{
		Name:          "datetime",
		BlueprintType: "string",
		ReclineType:   "date-time",
		DataTypes:     []string{"timestamp"},
	},
	{
		Name:          "email",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{"email"},
		Conformations: []string{"email"},
		DataTypes:     []string{"varchar(100)"},
	},
	{
		Name:          "namespace",
		BlueprintType: "string",
		ReclineType:   "string",
		DataTypes:     []string{"varchar(200)"},
	},
	{
		Name:          "name",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{"required"},
		Conformations: []string{"name"},
		DataTypes:     []string{"varchar(100)"},
	},
	{
		Name:          "encrypted",
		ReclineType:   "string",
		BlueprintType: "string",
		DataTypes:     []string{"varchar(100)", "varchar(500)", "varchar(500)", "text"},
	},
	{
		Name:          "json",
		ReclineType:   "string",
		BlueprintType: "string",
		DataTypes:     []string{"text", "varchar(100)"},
	},
	{
		Name:          "password",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{"required"},
		DataTypes:     []string{"varchar(200)"},
	},
	{
		Name:          "value",
		ReclineType:   "string",
		BlueprintType: "number",
		DataTypes:     []string{"varchar(100)"},
	},
	{
		Name:          "truefalse",
		BlueprintType: "boolean",
		ReclineType:   "boolean",
		DataTypes:     []string{"boolean"},
	},
	{
		Name:          "timestamp",
		BlueprintType: "timestamp",
		ReclineType:   "date-time",
		DataTypes:     []string{"timestamp"},
	},
	{
		Name:          "location",
		BlueprintType: "string",
		ReclineType:   "geo_point",
		DataTypes:     []string{"varchar(50)"},
	},
	{
		Name:          "location.latitude",
		BlueprintType: "number",
		ReclineType:   "number",
		Validations:   []string{"latitude"},
		DataTypes:     []string{"float(7,4)"},
	},
	{
		Name:          "location.longitude",
		BlueprintType: "number",
		ReclineType:   "number",
		Validations:   []string{"longitude"},
		DataTypes:     []string{"float(7,4)"},
	},
	{
		Name:          "location.altitude",
		BlueprintType: "string",
		ReclineType:   "number",
		DataTypes:     []string{"float(7,4)"},
	},
	{
		Name:          "color",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{"iscolor"},
		DataTypes:     []string{"varchar(50)"},
	},
	{
		Name:          "rating.10",
		BlueprintType: "number",
		ReclineType:   "string",
		Validations:   []string{"min=0,max=10"},
		DataTypes:     []string{"int(4)"},
	},
	{
		Name:          "measurement",
		ReclineType:   "number",
		BlueprintType: "number",
		DataTypes:     []string{"int(10)"},
	},
	{
		Name:          "label",
		ReclineType:   "string",
		BlueprintType: "string",
		DataTypes:     []string{"varchar(100)"},
	},
	{
		Name:          "content",
		ReclineType:   "string",
		BlueprintType: "string",
		DataTypes:     []string{"text"},
	},
	{
		Name:          "file",
		BlueprintType: "string",
		ReclineType:   "binary",
		Validations:   []string{"base64"},
		DataTypes:     []string{"text"},
	},
	{
		Name:          "url",
		BlueprintType: "string",
		ReclineType:   "string",
		Validations:   []string{"url"},
		DataTypes:     []string{"varchar(500)"},
	},
	{
		Name:          "image",
		BlueprintType: "string",
		ReclineType:   "binary",
		Validations:   []string{"base64"},
		DataTypes:     []string{"text"},
	},
}

type ColumnTypeManager struct {
	ColumnMap map[string]ColumnType
}

var ColumnManager *ColumnTypeManager

func InitialiseColumnManager() {
	ColumnManager = &ColumnTypeManager{}
	ColumnManager.ColumnMap = make(map[string]ColumnType)
	for _, col := range ColumnTypes {
		ColumnManager.ColumnMap[col.Name] = col
	}
}

func (ctm *ColumnTypeManager) GetBlueprintType(colName string) string {
	return ctm.ColumnMap[colName].BlueprintType
}

func (ctm *ColumnTypeManager) GetFakedata(colTypeName string) string {
	return fmt.Sprintf("%v", ctm.ColumnMap[colTypeName].Fake())
}

func (ctm *ColumnTypeManager) IsValidValue(val string, colType string, validator *validator2.Validate) error {
	if ctm.ColumnMap[colType].Validations == nil || len(ctm.ColumnMap[colType].Validations) < 1 {
		return nil
	}
	return validator.Var(val, ctm.ColumnMap[colType].Validations[0])

}

var CollectionTypes = []string{
	"Pair",
	"Triplet",
	"Set",
	"OrderedSet",
}
