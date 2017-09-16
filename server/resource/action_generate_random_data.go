package resource

import (
	log "github.com/sirupsen/logrus"
	"github.com/artpar/api2go"
	"github.com/satori/go.uuid"
	"github.com/artpar/goms/server/auth"
	"net/http"
	"context"
)

type RandomDataGeneratePerformer struct {
	cmsConfig *CmsConfig
	cruds     map[string]*DbResource
	tableMap  map[string][]api2go.ColumnInfo
}

func (d *RandomDataGeneratePerformer) Name() string {
	return "generate.random.data"
}

func (d *RandomDataGeneratePerformer) DoAction(request ActionRequest, inFields map[string]interface{}) ([]ActionResponse, []error) {

	responses := make([]ActionResponse, 0)

	subjectInstance := inFields["subject"].(map[string]interface{})
	user := inFields["user"]
	userReferenceId := ""
	//userIdInt := uint64(1)
	var err error

	userMap := user.(map[string]interface{})
	userReferenceId = userMap["reference_id"].(string)
	userIdInt := userMap["id"].(int64)
	//userIdInt, err = d.cruds["user"].GetReferenceIdToId("user", userReferenceId)
	if err != nil {
		log.Errorf("Failed to get user id from user reference id: %v", err)
	}
	tableName := subjectInstance["table_name"].(string)

	count := int(inFields["count"].(float64))

	rows := make([]map[string]interface{}, 0)
	for i := 0; i < count; i++ {
		row := GetFakeRow(d.tableMap[tableName])
		row["reference_id"] = uuid.NewV4().String()
		row["permission"] = auth.DEFAULT_PERMISSION
		rows = append(rows, row)
	}

	httpRequest := &http.Request{
		Method: "POST",
	}
	httpRequest = httpRequest.WithContext(context.WithValue(context.Background(), "user_id", userReferenceId))
	httpRequest = httpRequest.WithContext(context.WithValue(httpRequest.Context(), "user_id_integer", int64(userIdInt)))
	httpRequest = httpRequest.WithContext(context.WithValue(httpRequest.Context(), "usergroup_id", []auth.GroupPermission{}))

	req := api2go.Request{
		PlainRequest: httpRequest,
	}
	for _, row := range rows {

		_, err := d.cruds[tableName].Create(api2go.NewApi2GoModelWithData(tableName, nil, 0, nil, row), req)
		if err != nil {
			log.Errorf("Was about to insert this fake object: %v", row)
			log.Errorf("Failed to direct insert into table [%v] : %v", tableName, err)
		}
	}
	return responses, nil
}

func GetFakeRow(columns []api2go.ColumnInfo) map[string]interface{} {

	row := make(map[string]interface{})

	for _, col := range columns {

		if col.IsForeignKey {
			continue
		}

		isStandardColumn := false
		for _, c := range StandardColumns {
			if col.ColumnName == c.ColumnName {
				isStandardColumn = true
			}
		}

		if (isStandardColumn) {
			continue
		}

		fakeValue := ColumnManager.GetFakedata(col.ColumnType)

		row[col.ColumnName] = fakeValue

	}

	return row

}

func NewRandomDataGeneratePerformer(initConfig *CmsConfig, cruds map[string]*DbResource) (ActionPerformerInterface, error) {

	tableMap := make(map[string][]api2go.ColumnInfo)
	for _, table := range initConfig.Tables {
		tableMap[table.TableName] = table.Columns
	}

	handler := RandomDataGeneratePerformer{
		cmsConfig: initConfig,
		cruds:     cruds,
		tableMap:  tableMap,
	}

	return &handler, nil

}
