package resource

import (
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"strings"
	"encoding/json"
)

type ImportDataPerformer struct {
	cmsConfig *CmsConfig
	cruds     map[string]*DbResource
}

func (d *ImportDataPerformer) Name() string {
	return "__data_import"
}

func (d *ImportDataPerformer) DoAction(request ActionRequest, inFields map[string]interface{}) ([]ActionResponse, []error) {

	responses := make([]ActionResponse, 0)

	subjectInstance, isSubjected := inFields["subject"]
	user, isUserPresent := inFields["user"]
	userReferenceId := ""
	userIdInt := uint64(1)
	var err error
	if isUserPresent {
		userMap := user.(map[string]interface{})
		userReferenceId = userMap["reference_id"].(string)
		userIdInt, err = d.cruds["user"].GetReferenceIdToId("user", userReferenceId)
		if err != nil {
			log.Errorf("Failed to get user id from user reference id: %v", err)
		}
	}

	files := inFields["dump_file"].([]interface{})

	truncate_before_insert := inFields["truncate_before_insert"].(bool)
	//execute_middleware_chain := inFields["execute_middleware_chain"].(bool)

	imports := make(map[string][]interface{})

	for _, fileInterface := range files {
		file := fileInterface.(map[string]interface{})

		fileName := file["name"].(string)
		fileContentsBase64 := file["file"].(string)
		fileBytes, err := base64.StdEncoding.DecodeString(strings.Split(fileContentsBase64, ",")[1])

		if err != nil {
			log.Errorf("Failed to read file contents as base64 encoded: %v", err)
			continue
		}

		log.Infof("Processing file: %v", fileName)

		var jsonData map[string]interface{}

		err = json.Unmarshal(fileBytes, &jsonData)
		if err != nil {
			log.Errorf("Failed to read data as list: %v", err)
			continue
		}

		if isSubjected {
			subjectInstanceMap := subjectInstance.(map[string]interface{})
			subjectTableName := subjectInstanceMap["table_name"].(string)

			_, ok := imports[subjectTableName]
			if !ok {
				arr := make([]interface{}, 0)
				imports[subjectTableName] = arr
			}
			imports[subjectTableName] = append(imports[subjectTableName], jsonData[subjectTableName])
		} else {
			for tableName, val := range jsonData {

				_, ok := imports[tableName]
				if !ok {
					arr := make([]interface{}, 0)
					imports[tableName] = arr
				}
				imports[tableName] = append(imports[tableName], val)

			}

		}

	}

	for tableName, importedDatas := range imports {

		if truncate_before_insert {
			err := d.cruds[tableName].TruncateTable(tableName)
			if err != nil {
				log.Errorf("Failed to truncate table before importing data: %v", err)
			}
		}

		for _, importedData := range importedDatas {
			dataAsArray, ok := importedData.([]interface{})
			if !ok {
				log.Errorf("Data for [%v] in invalid format", tableName)
				continue
			}

			for _, row := range dataAsArray {
				data := row.(map[string]interface{})

				if isUserPresent {
					data["user_id"] = userIdInt
				}

				err := d.cruds[tableName].DirectInsert(tableName, data)
				if err != nil {
					log.Errorf("Was about to inser this: %v", data)
					log.Errorf("Failed to direct insert into table [%v] : %v", tableName, err)
				}
			}
		}
	}
	return responses, nil
}

func NewImportDataPerformer(initConfig *CmsConfig, cruds map[string]*DbResource) (ActionPerformerInterface, error) {

	handler := ImportDataPerformer{
		cmsConfig: initConfig,
		cruds:     cruds,
	}

	return &handler, nil

}
