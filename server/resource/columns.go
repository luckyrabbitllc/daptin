package resource

import (
	"github.com/artpar/api2go"
)

var StandardColumns = []api2go.ColumnInfo{
	{
		Name:            "id",
		ColumnName:      "id",
		DataType:        "INTEGER",
		IsPrimaryKey:    true,
		IsAutoIncrement: true,
		ExcludeFromApi:  true,
		ColumnType:      "id",
	},
	{
		Name:           "version",
		ColumnName:     "version",
		DataType:       "INTEGER",
		ColumnType:     "measurement",
		DefaultValue:   "1",
		ExcludeFromApi: true,
	},
	{
		Name:         "created_at",
		ColumnName:   "created_at",
		DataType:     "timestamp",
		DefaultValue: "current_timestamp",
		ColumnType:   "datetime",
		IsIndexed:    true,
	},
	{
		Name:       "updated_at",
		ColumnName: "updated_at",
		DataType:   "timestamp",
		IsIndexed:  true,
		IsNullable: true,
		ColumnType: "datetime",
	},
	{
		Name:       "reference_id",
		ColumnName: "reference_id",
		DataType:   "varchar(40)",
		IsIndexed:  true,
		ColumnType: "alias",
	},
	{
		Name:       "permission",
		ColumnName: "permission",
		DataType:   "int(11)",
		IsIndexed:  false,
		ColumnType: "value",
	},
}

var StandardRelations = []api2go.TableRelation{
	api2go.NewTableRelation("world_column", "belongs_to", "world"),
	api2go.NewTableRelation("action", "belongs_to", "world"),
	api2go.NewTableRelation("world", "has_many", "smd"),
	api2go.NewTableRelation("oauth_token", "has_one", "oauth_connect"),
	api2go.NewTableRelation("data_exchange", "has_one", "oauth_token"),
	api2go.NewTableRelation("timeline", "belongs_to", "world"),
	api2go.NewTableRelation("cloud_store", "has_one", "oauth_token"),
	api2go.NewTableRelation("site", "has_one", "cloud_store"),
}

var SystemSmds = []LoopbookFsmDescription{}
var SystemExchanges = []ExchangeContract{}

var SystemActions = []Action{
	{
		Name: "restart_goms",
		Label: "Restart system",
		OnType: "world",
		InstanceOptional: true,
		InFields: []api2go.ColumnInfo{

		},
		OutFields: []Outcome {
			{
				Type:   "system_json_schema_update",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"json_schema": "!JSON.parse('[{\"name\":\"empty.json\",\"file\":\"data:application/json;base64,e30K\",\"type\":\"application/json\"}]')",
				},
			},
		},
	},
	{
		Name:             "publish_package_to_market",
		Label:            "Update package list",
		OnType:           "marketplace",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{

		},
		OutFields: []Outcome{
			{
				Type:   "market.package.refresh",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{

				},
			},
		},
	},
	{
		Name:             "visit_marketplace_github",
		Label:            "Go to marketplace",
		OnType:           "marketplace",
		InstanceOptional: false,
		InFields:         []api2go.ColumnInfo{},
		OutFields: []Outcome{
			{
				Type:   "client.redirect",
				Method: "ACTIONRESPONSE",
				Attributes: map[string]interface{}{
					"location": "$subject.endpoint",
					"window":   "_blank",
				},},
		},
	},
	{
		Name:             "refresh_marketplace_packages",
		Label:            "Refresh marketplace",
		OnType:           "marketplace",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{
		},
		OutFields: []Outcome{
			{
				Type:   "marketplace.package.refresh",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
				},
			},
		},
	},
	{
		Name:             "generate_random_data",
		Label:            "Generate random data",
		OnType:           "world",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "Number of records",
				ColumnName: "count",
				ColumnType: "measurement",
			},
		},
		OutFields: []Outcome{
			{
				Type:   "generate.random.data",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"count": "~count",
				},
			},
		},
		Validations: []ColumnTag{
			{
				ColumnName: "count",
				Tags:       "gt=0",
			},
		},
	},
	//{
	//
	//	Name: "update_config",
	//	Label: "Update configuration",
	//	OnType: "world",
	//	InstanceOptional: true,
	//	InFields: []api2go.ColumnInfo{
	//		{
	//			Name: "default_storage",
	//		},
	//	},
	//},
	{
		Name:             "install_marketplace_package",
		Label:            "Install package from market",
		OnType:           "marketplace",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "package_name",
				ColumnName: "package_name",
				ColumnType: "label",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "marketplace.package.install",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"package_name":        "~package_name",
					"market_reference_id": "$.reference_id",
				},
			},
		},
	},
	{
		Name:             "export_data",
		Label:            "Export data for backup",
		OnType:           "world",
		InstanceOptional: true,
		InFields: []api2go.ColumnInfo{
		},
		OutFields: []Outcome{
			{
				Type:   "__data_export",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"world_reference_id": "$.reference_id",
				},
			},
		},
	},
	{
		Name:             "import_data",
		Label:            "Import data from dump",
		OnType:           "world",
		InstanceOptional: true,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "JSON Dump file",
				ColumnName: "dump_file",
				ColumnType: "file.json|yaml|toml|hcl",
				IsNullable: false,
			},
			{
				Name:       "truncate_before_insert",
				ColumnName: "truncate_before_insert",
				ColumnType: "truefalse",
			},
		},
		OutFields: []Outcome{
			{
				Type:   "__data_import",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"world_reference_id":       "$.reference_id",
					"execute_middleware_chain": "~execute_middleware_chain",
					"truncate_before_insert":   "~truncate_before_insert",
					"dump_file":                "~dump_file",
				},
			},
		},
	},
	{
		Name:             "upload_file",
		Label:            "Upload file to external store",
		OnType:           "cloud_store",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "File",
				ColumnName: "file",
				ColumnType: "file.*",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "__external_file_upload",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"file": "~file",
				},
			},
		},
	},
	{
		Name:             "upload_system_schema",
		Label:            "Upload features",
		OnType:           "world",
		InstanceOptional: true,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "Schema file",
				ColumnName: "schema_file",
				ColumnType: "file.json|yaml|toml|hcl",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "system_json_schema_update",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"json_schema": "~schema_file",
				},
			},
		},
	},
	{
		Name:             "upload_xls_to_system_schema",
		Label:            "Upload xls to entity",
		OnType:           "world",
		InstanceOptional: true,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "XLSX file",
				ColumnName: "data_xls_file",
				ColumnType: "file.application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				IsNullable: false,
			},
			{
				Name:       "Entity name",
				ColumnName: "entity_name",
				ColumnType: "label",
				IsNullable: false,
			},
		},
		Validations: []ColumnTag{
			{
				ColumnName: "entity_name",
				Tags:       "required",
			},
		},
		OutFields: []Outcome{
			{
				Type:   "__upload_file_to_entity",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"xls_data": "~data_xls_file",
					"name":     "entity_name",
				},
			},
		},
	},
	{
		Name:             "download_system_schema",
		Label:            "Download system schema",
		OnType:           "world",
		InstanceOptional: true,
		InFields:         []api2go.ColumnInfo{},
		OutFields: []Outcome{
			{
				Type:       "__download_cms_config",
				Method:     "EXECUTE",
				Attributes: map[string]interface{}{},
			},
		},
	},
	{
		Name:             "invoke_become_admin",
		Label:            "Become GoMS admin",
		InstanceOptional: true,
		OnType:           "world",
		InFields:         []api2go.ColumnInfo{},
		OutFields: []Outcome{
			{
				Type:   "__become_admin",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"user_id": "$user.id",
				},
			},
		},
	},
	{
		Name:             "signup",
		Label:            "Sign up",
		InstanceOptional: true,
		OnType:           "user",
		InFields: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				ColumnType: "label",
				IsNullable: false,
			},
			{
				Name:       "email",
				ColumnName: "email",
				ColumnType: "email",
				IsNullable: false,
			},
			{
				Name:       "password",
				ColumnName: "password",
				ColumnType: "password",
				IsNullable: false,
			},
			{
				Name:       "Password Confirm",
				ColumnName: "passwordConfirm",
				ColumnType: "password",
				IsNullable: false,
			},
		},
		Validations: []ColumnTag{
			{
				ColumnName: "email",
				Tags:       "email",
			},
			{
				ColumnName: "name",
				Tags:       "required",
			},
			{
				ColumnName: "password",
				Tags:       "eqfield=InnerStructField[passwordConfirm],min=8",
			},
		},
		Conformations: []ColumnTag{
			{
				ColumnName: "email",
				Tags:       "email",
			},
			{
				ColumnName: "name",
				Tags:       "trim",
			},
		},
		OutFields: []Outcome{
			{
				Type:      "user",
				Method:    "POST",
				Reference: "user",
				Attributes: map[string]interface{}{
					"name":      "~name",
					"email":     "~email",
					"password":  "~password",
					"confirmed": "0",
				},
			},
			{
				Type:      "usergroup",
				Method:    "POST",
				Reference: "usergroup",
				Attributes: map[string]interface{}{
					"name": "!'Home group for ' + user.name",
				},
			},
			{
				Type:      "user_user_id_has_usergroup_usergroup_id",
				Method:    "POST",
				Reference: "user_usergroup",
				Attributes: map[string]interface{}{
					"user_id":      "$user.reference_id",
					"usergroup_id": "$usergroup.reference_id",
				},
			},
			{
				Type:   "client.notify",
				Method: "ACTIONRESPONSE",
				Attributes: map[string]interface{}{
					"type":    "success",
					"title":   "Success",
					"message": "Signup Successful",
				},
			},
			{
				Type:   "client.redirect",
				Method: "ACTIONRESPONSE",
				Attributes: map[string]interface{}{
					"location": "/auth/signin",
					"window":   "self",
				},
			},
		},
	},
	{
		Name:             "signin",
		Label:            "Sign in",
		InstanceOptional: true,
		OnType:           "user",
		InFields: []api2go.ColumnInfo{
			{
				Name:       "email",
				ColumnName: "email",
				ColumnType: "email",
				IsNullable: false,
			},
			{
				Name:       "password",
				ColumnName: "password",
				ColumnType: "password",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "jwt.token",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"email":    "~email",
					"password": "~password",
				},
			},
		},
	},
	{
		Name:   "oauth.login.begin",
		Label:  "Authenticate via OAuth",
		OnType: "oauth_connect",
		InFields: []api2go.ColumnInfo{
		},
		OutFields: []Outcome{
			{
				Type:   "oauth.client.redirect",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"authenticator": "$.name",
					"scope":         "$.scope",
				},
			},
		},
	},
	{
		Name:             "oauth.login.response",
		Label:            "",
		InstanceOptional: true,
		OnType:           "oauth_token",
		InFields: []api2go.ColumnInfo{
			{
				Name:       "code",
				ColumnName: "code",
				ColumnType: "hidden",
				IsNullable: false,
			},
			{
				Name:       "state",
				ColumnName: "state",
				ColumnType: "hidden",
				IsNullable: false,
			},
			{
				Name:       "authenticator",
				ColumnName: "authenticator",
				ColumnType: "hidden",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "oauth.login.response",
				Method: "EXECUTE",
				Attributes: map[string]interface{}{
					"authenticator": "~authenticator",
				},
			},
		},
	},
	{
		Name:             "add_exchange",
		Label:            "Add new data exchange",
		OnType:           "world",
		InstanceOptional: false,
		InFields: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				ColumnType: "name",
				IsNullable: false,
			},
			{
				Name:       "sheet_id",
				ColumnName: "sheet_id",
				ColumnType: "alias",
				IsNullable: false,
			},
			{
				Name:       "app_key Key",
				ColumnName: "app_key",
				ColumnType: "alias",
				IsNullable: false,
			},
		},
		OutFields: []Outcome{
			{
				Type:   "data_exchange",
				Method: "POST",
				Attributes: map[string]interface{}{
					"name":              "!'Export ' + subject.table_name + ' to excel sheet'",
					"source_attributes": "!JSON.stringify({name: subject.table_name})",
					"source_type":       "self",
					"target_type":       "gsheet-append",
					"options":           "!JSON.stringify({hasHeader: true})",
					"attributes":        "!JSON.stringify([{SourceColumn: '$self.description', TargetColumn: 'Task description'}])",
					"target_attributes": "!JSON.stringify({sheetUrl: 'https://content-sheets.googleapis.com/v4/spreadsheets/' + sheet_id + '/values/A1:append', appKey: app_key})",
				},
			},
		},
	},
}

var StandardTables = []TableInfo{
	{
		TableName: "marketplace",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				DataType:   "varchar(100)",
				ColumnType: "label",
				IsIndexed:  true,
			},
			{
				Name:       "endpoint",
				ColumnName: "endpoint",
				DataType:   "varchar(200)",
				ColumnType: "url",
			},
			{
				Name:         "root_path",
				ColumnName:   "root_path",
				DataType:     "varchar(100)",
				ColumnType:   "label",
				DefaultValue: "''",
			},
		},
	},
	{
		TableName: "json_schema",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "schema_name",
				ColumnName: "schema_name",
				ColumnType: "label",
				DataType:   "varchar(100)",
				IsNullable: false,
			},
			{
				Name:       "json_schema",
				ColumnType: "json",
				DataType:   "text",
				ColumnName: "json_schema",
			},
		},
	},
	{
		TableName: "timeline",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "event_type",
				ColumnName: "event_type",
				ColumnType: "label",
				DataType:   "varchar(50)",
				IsNullable: false,
			},
			{
				Name:       "title",
				ColumnName: "title",
				ColumnType: "label",
				IsIndexed:  true,
				DataType:   "varchar(50)",
				IsNullable: false,
			},
			{
				Name:       "payload",
				ColumnName: "payload",
				ColumnType: "content",
				DataType:   "text",
				IsNullable: true,
			},
		},
	},
	{
		TableName: "world",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "table_name",
				ColumnName: "table_name",
				IsNullable: false,
				IsUnique:   true,
				IsIndexed:  true,
				DataType:   "varchar(200)",
				ColumnType: "name",
			},
			{
				Name:       "world_schema_json",
				ColumnName: "world_schema_json",
				DataType:   "text",
				IsNullable: false,
				ColumnType: "json",
			},
			{
				Name:         "default_permission",
				ColumnName:   "default_permission",
				DataType:     "int(4)",
				IsNullable:   false,
				DefaultValue: "644",
				ColumnType:   "value",
			},

			{
				Name:         "is_top_level",
				ColumnName:   "is_top_level",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "true",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_hidden",
				ColumnName:   "is_hidden",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_join_table",
				ColumnName:   "is_join_table",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_state_tracking_enabled",
				ColumnName:   "is_state_tracking_enabled",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
		},
	},
	{
		TableName: "world_column",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				DataType:   "varchar(100)",
				IsIndexed:  true,
				IsNullable: false,
				ColumnType: "name",
			},
			{
				Name:       "column_name",
				ColumnName: "column_name",
				DataType:   "varchar(100)",
				IsIndexed:  true,
				IsNullable: false,
				ColumnType: "name",
			},
			{
				Name:       "column_type",
				ColumnName: "column_type",
				DataType:   "varchar(100)",
				IsNullable: false,
				ColumnType: "label",
			},
			{
				Name:       "column_description",
				ColumnName: "column_description",
				DataType:   "varchar(100)",
				IsNullable: true,
				ColumnType: "content",
			},
			{
				Name:         "is_primary_key",
				ColumnName:   "is_primary_key",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_auto_increment",
				ColumnName:   "is_auto_increment",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_indexed",
				ColumnName:   "is_indexed",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_unique",
				ColumnName:   "is_unique",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_nullable",
				ColumnName:   "is_nullable",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "is_foreign_key",
				ColumnName:   "is_foreign_key",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "false",
				ColumnType:   "truefalse",
			},
			{
				Name:         "include_in_api",
				ColumnName:   "include_in_api",
				DataType:     "bool",
				IsNullable:   false,
				DefaultValue: "true",
				ColumnType:   "truefalse",
			},
			{
				Name:       "foreign_key_data",
				ColumnName: "foreign_key_data",
				DataType:   "varchar(100)",
				IsNullable: true,
				ColumnType: "content",
			},
			{
				Name:       "default_value",
				ColumnName: "default_value",
				DataType:   "varchar(100)",
				IsNullable: true,
				ColumnType: "content",
			},
			{
				Name:       "data_type",
				ColumnName: "data_type",
				DataType:   "varchar(50)",
				IsNullable: true,
				ColumnType: "label",
			},
		},
	},
	{
		TableName: "stream",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "stream_name",
				ColumnName: "stream_name",
				DataType:   "varchar(100)",
				IsNullable: false,
				ColumnType: "label",
				IsIndexed:  true,
			},
			{
				Name:       "stream_contract",
				ColumnName: "stream_contract",
				DataType:   "text",
				IsNullable: false,
				ColumnType: "json",
			},
		},
	},
	{
		TableName: "user",
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				IsIndexed:  true,
				DataType:   "varchar(80)",
				ColumnType: "name",
			},
			{
				Name:       "email",
				ColumnName: "email",
				DataType:   "varchar(80)",
				IsIndexed:  true,
				IsUnique:   true,
				ColumnType: "email",
			},

			{
				Name:       "password",
				ColumnName: "password",
				DataType:   "varchar(100)",
				ColumnType: "password",
				IsNullable: true,
			},
			{
				Name:         "confirmed",
				ColumnName:   "confirmed",
				DataType:     "boolean",
				ColumnType:   "truefalse",
				IsNullable:   false,
				DefaultValue: "false",
			},
		},
		Validations: []ColumnTag{
			{
				ColumnName: "email",
				Tags:       "email",
			},
			{
				ColumnName: "password",
				Tags:       "required",
			},
			{
				ColumnName: "name",
				Tags:       "required",
			},
		},
		Conformations: []ColumnTag{
			{
				ColumnName: "email",
				Tags:       "email",
			},
		},
	},
	{
		TableName: "usergroup",
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				IsIndexed:  true,
				DataType:   "varchar(80)",
				ColumnType: "name",
			},
		},
	},
	{
		TableName: "action",
		Columns: []api2go.ColumnInfo{
			{
				Name:       "action_name",
				IsIndexed:  true,
				ColumnName: "action_name",
				DataType:   "varchar(100)",
				ColumnType: "name",
			},
			{
				Name:       "label",
				ColumnName: "label",
				IsIndexed:  true,
				DataType:   "varchar(100)",
				ColumnType: "label",
			},
			{
				Name:         "instance_optional",
				ColumnName:   "instance_optional",
				IsIndexed:    false,
				DataType:     "bool",
				ColumnType:   "truefalse",
				DefaultValue: "true",
			},
			{
				Name:       "action_schema",
				ColumnName: "action_schema",
				DataType:   "text",
				ColumnType: "json",
			},
		},
	},
	{
		TableName: "smd",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				IsIndexed:  true,
				DataType:   "varchar(100)",
				ColumnType: "label",
				IsNullable: false,
			},
			{
				Name:       "label",
				ColumnName: "label",
				DataType:   "varchar(100)",
				ColumnType: "label",
				IsNullable: false,
			},
			{
				Name:       "initial_state",
				ColumnName: "initial_state",
				DataType:   "varchar(100)",
				ColumnType: "label",
				IsNullable: false,
			},
			{
				Name:       "events",
				ColumnName: "events",
				DataType:   "text",
				ColumnType: "json",
				IsNullable: false,
			},
		},
	},
	{
		TableName: "oauth_connect",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				IsUnique:   true,
				IsIndexed:  true,
				DataType:   "varchar(80)",
				ColumnType: "name",
			},
			{
				Name:       "client_id",
				ColumnName: "client_id",
				DataType:   "varchar(80)",
				ColumnType: "name",
			},
			{
				Name:       "client_secret",
				ColumnName: "client_secret",
				DataType:   "varchar(80)",
				ColumnType: "encrypted",
			},
			{
				Name:         "scope",
				ColumnName:   "scope",
				DataType:     "varchar(1000)",
				ColumnType:   "content",
				DefaultValue: "'https://www.googleapis.com/auth/spreadsheets'",
			},
			{
				Name:         "response_type",
				ColumnName:   "response_type",
				DataType:     "varchar(80)",
				ColumnType:   "name",
				DefaultValue: "'code'",
			},
			{
				Name:       "redirect_uri",
				ColumnName: "redirect_uri",
				DataType:   "varchar(80)",
				ColumnType: "url",
				DefaultValue: "'https://dashboard.devsupport.ai/oauth/response'",
			},
			{
				Name:         "auth_url",
				ColumnName:   "auth_url",
				DataType:     "varchar(200)",
				DefaultValue: "'https://accounts.google.com/o/oauth2/auth'",
				ColumnType:   "url",
			},
			{
				Name:         "token_url",
				ColumnName:   "token_url",
				DataType:     "varchar(200)",
				DefaultValue: "'https://accounts.google.com/o/oauth2/token'",
				ColumnType:   "url",
			},
		},
	},
	{
		TableName: "data_exchange",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				ColumnType: "name",
				DataType:   "varchar(200)",
				IsIndexed:  true,
			},
			{
				Name:       "source_attributes",
				ColumnName: "source_attributes",
				ColumnType: "json",
				DataType:   "text",
			},
			{
				Name:       "source_type",
				ColumnName: "source_type",
				ColumnType: "name",
				DataType:   "varchar(100)",
			},
			{
				Name:       "target_attributes",
				ColumnName: "target_attributes",
				ColumnType: "json",
				DataType:   "text",
			},
			{
				Name:       "target_type",
				ColumnName: "target_type",
				ColumnType: "name",
				DataType:   "varchar(100)",
			},
			{
				Name:       "attributes",
				ColumnName: "attributes",
				ColumnType: "json",
				DataType:   "text",
			},
			{
				Name:       "options",
				ColumnName: "options",
				ColumnType: "json",
				DataType:   "text",
			},
		},
	},
	{
		TableName: "oauth_token",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "access_token",
				ColumnName: "access_token",
				ColumnType: "encrypted",
				DataType:   "varchar(1000)",
			},
			{
				Name:       "expires_in",
				ColumnName: "expires_in",
				ColumnType: "measurement",
				DataType:   "int(11)",
			},
			{
				Name:       "refresh_token",
				ColumnName: "refresh_token",
				ColumnType: "encrypted",
				DataType:   "varchar(1000)",
			},
			{
				Name:       "token_type",
				ColumnName: "token_type",
				ColumnType: "label",
				DataType:   "varchar(20)",
			},
		},
	},
	{
		TableName: "cloud_store",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "Name",
				ColumnName: "name",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
			{
				Name:       "store_type",
				ColumnName: "store_type",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
			{
				Name:       "store_provider",
				ColumnName: "store_provider",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
			{
				Name:       "root_path",
				ColumnName: "root_path",
				ColumnType: "label",
				DataType:   "varchar(1000)",
			},
			{
				Name:       "store_parameters",
				ColumnName: "store_parameters",
				ColumnType: "json",
				DataType:   "text",
			},
		},
	},
	{
		TableName: "site",
		IsHidden:  true,
		Columns: []api2go.ColumnInfo{
			{
				Name:       "name",
				ColumnName: "name",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
			{
				Name:       "hostname",
				ColumnName: "hostname",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
			{
				Name:       "path",
				ColumnName: "path",
				ColumnType: "label",
				DataType:   "varchar(100)",
			},
		},
	},
}

var StandardMarketplaces = []Marketplace{

}

var StandardStreams = []StreamContract{
	{
		StreamName:     "table",
		RootEntityName: "world",
		Columns: []api2go.ColumnInfo{
			{
				Name:       "table_name",
				ColumnType: "label",
			},
			{
				Name:       "reference_id",
				ColumnType: "label",
			},
		},
	},
	{
		StreamName:     "transformed_user",
		RootEntityName: "user",
		Columns: []api2go.ColumnInfo{
			{
				Name:       "transformed_user_name",
				ColumnType: "label",
			},
			{
				Name:       "primary_email",
				ColumnType: "label",
			},
		},
		Transformations: []Transformation{
			{
				Operation: "select",
				Attributes: map[string]interface{}{
					"columns": []string{"name", "email"},
				},
			},
			{
				Operation: "rename",
				Attributes: map[string]interface{}{
					"oldName": "name",
					"newName": "transformed_user_name",
				},
			},
			{
				Operation: "rename",
				Attributes: map[string]interface{}{
					"oldName": "email",
					"newName": "primary_email",
				},
			},
		},
	},
}

type TableInfo struct {
	TableName              string `db:"table_name"`
	TableId                int
	DefaultPermission      int64  `db:"default_permission"`
	Columns                []api2go.ColumnInfo
	StateMachines          []LoopbookFsmDescription
	Relations              []api2go.TableRelation
	IsTopLevel             bool   `db:"is_top_level"`
	Permission             int64
	UserId                 uint64 `db:"user_id"`
	IsHidden               bool   `db:"is_hidden"`
	IsJoinTable            bool   `db:"is_join_table"`
	IsStateTrackingEnabled bool   `db:"is_state_tracking_enabled"`
	IsAuditEnabled         bool   `db:"is_audit_enabled"`
	Validations            []ColumnTag
	Conformations          []ColumnTag
}

type ColumnTag struct {
	ColumnName string
	Tags       string
}
