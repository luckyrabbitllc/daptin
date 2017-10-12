package resource

import (
	"fmt"
	"github.com/artpar/api2go"
	"github.com/artpar/daptin/server/auth"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"strings"
)

func CreateUniqueConstraints(initConfig *CmsConfig, db *sqlx.DB) {
	for _, table := range initConfig.Tables {

		for _, column := range table.Columns {

			if column.IsUnique {
				indexName := "index_" + table.TableName + "_" + column.ColumnName + "_unique"
				alterTable := "create unique index " + indexName + " on " + table.TableName + "(" + column.ColumnName + ")"
				log.Infof("Create unique index sql: %v", alterTable)
				_, err := db.Exec(alterTable)
				if err != nil {
					//log.Infof("Table[%v] Column[%v]: Failed to create unique index: %v", table.TableName, column.ColumnName, err)
				}
			}
		}

		if strings.Index(table.TableName, "_has_") > -1 {

			cols := []string{}

			for _, col := range table.Columns {
				if col.IsForeignKey {
					cols = append(cols, col.ColumnName)
				}
			}

			indexName := GetMD5Hash("index_join_" + table.TableName + "_" + "_unique")
			alterTable := "create unique index " + indexName + " on " + table.TableName + "(" + strings.Join(cols, ", ") + ")"
			log.Infof("Create unique index sql: %v", alterTable)
			_, err := db.Exec(alterTable)
			if err != nil {
				log.Infof("Table[%v] Column[%v]: Failed to create unique join index: %v", table.TableName, err)
			}

		}

	}
}

func CreateIndexes(initConfig *CmsConfig, db *sqlx.DB) {
	for _, table := range initConfig.Tables {
		for _, column := range table.Columns {

			if column.IsUnique {
				indexName := "u" + GetMD5Hash("index_"+table.TableName+"_"+column.ColumnName+"_index")
				alterTable := "create unique index " + indexName + " on " + table.TableName + " (" + column.ColumnName + ")"
				//log.Infof("Create index sql: %v", alterTable)
				_, err := db.Exec(alterTable)
				if err != nil {
					//log.Infof("Failed to create index on Table[%v] Column[%v]: %v", table.TableName, column.ColumnName, err)
				}
			} else if column.IsIndexed {
				indexName := "i" + GetMD5Hash("index_"+table.TableName+"_"+column.ColumnName+"_index")
				alterTable := "create index " + indexName + " on " + table.TableName + " (" + column.ColumnName + ")"
				//log.Infof("Create index sql: %v", alterTable)
				_, err := db.Exec(alterTable)
				if err != nil {
					//log.Infof("Failed to create index on Table[%v] Column[%v]: %v", table.TableName, column.ColumnName, err)
				}
			}
		}
	}
}

func CreateRelations(initConfig *CmsConfig, db *sqlx.DB) {

	for i, table := range initConfig.Tables {
		for _, column := range table.Columns {
			if column.IsForeignKey {
				keyName := "fk" + GetMD5Hash(table.TableName+"_"+column.ColumnName+"_"+column.ForeignKeyData.TableName+"_"+column.ForeignKeyData.ColumnName+"_fk")

				if db.DriverName() == "sqlite3" {
					continue
				}

				alterSql := "alter table " + table.TableName + " add constraint " + keyName + " foreign key (" + column.ColumnName + ") references " + column.ForeignKeyData.String()
				//log.Infof("Alter table add constraint sql: %v", alterSql)
				_, err := db.Exec(alterSql)
				if err != nil {
					//log.Infof("Failed to create foreign key [%v], probably it exists: %v", err, keyName)
				} else {
					log.Infof("Key created [%v][%v]", keyName, table.TableName)
				}
			}
		}

		relations := make([]api2go.TableRelation, 0)

		for _, rel := range initConfig.Relations {
			if rel.GetSubject() == table.TableName || rel.GetObject() == table.TableName {
				relations = append(relations, rel)
			}
		}

		//initConfig.Tables[i].AddRelation(relations...)
		// reset relations
		initConfig.Tables[i].Relations = relations
	}
}

func CheckAuditTables(config *CmsConfig, db *sqlx.DB) {

	newRelations := make([]api2go.TableRelation, 0)

	tableMap := make(map[string]*TableInfo)
	for i := range config.Tables {
		t := config.Tables[i]
		tableMap[t.TableName] = &t
	}

	createAuditTableFor := make([]string, 0)
	updateAuditTableFor := make([]string, 0)

	for _, table := range config.Tables {

		if api2go.EndsWithCheck(table.TableName, "_audit") {
			log.Infof("[%v] is an audit table", table.TableName)
			continue
		}

		auditTableName := table.TableName + "_audit"
		existingAuditTable, ok := tableMap[auditTableName]
		if !ok {
			createAuditTableFor = append(createAuditTableFor, table.TableName)
		} else {

			if len(table.Columns) > len(existingAuditTable.Columns) {
				log.Infof("New columns added to the table, audit table need to be updated")
				updateAuditTableFor = append(updateAuditTableFor, table.TableName)
			}
		}

	}

	for _, tableName := range createAuditTableFor {

		table := tableMap[tableName]
		columnsCopy := make([]api2go.ColumnInfo, 0)
		auditTableName := tableName + "_audit"
		log.Infof("Create audit table [%s] for table [%v]", table.TableName, auditTableName)

		for _, col := range table.Columns {

			var c api2go.ColumnInfo
			err := copier.Copy(&c, &col)
			if err != nil {
				log.Errorf("Failed to copy columns for audit table: %v", err)
				continue
			}

			if c.ColumnName == "id" {
				continue
			}

			if c.ColumnType == "datetime" {
				c.IsNullable = true
			}

			c.IsUnique = false
			c.IsPrimaryKey = false
			c.IsAutoIncrement = false

			log.Infof("Add column to table [%v] == [%v]", auditTableName, c)
			columnsCopy = append(columnsCopy, c)

		}

		//newRelation := api2go.TableRelation{
		//	Subject:    auditTableName,
		//	Relation:   "belongs_to",
		//	Object:     tableName,
		//	ObjectName: "audit_object_id",
		//}

		//newRelations = append(newRelations, newRelation)

		newTable := TableInfo{
			TableName:         auditTableName,
			Columns:           columnsCopy,
			IsHidden:          true,
			DefaultPermission: auth.NewPermission(auth.Read, auth.Read, auth.Read).IntValue(),
			Permission:        auth.NewPermission(auth.Create, auth.Create, auth.Create).IntValue(),
		}

		config.Tables = append(config.Tables, newTable)
	}

	log.Infof("%d Audit tables are new", len(createAuditTableFor))
	log.Infof("%d Audit tables are updated", len(updateAuditTableFor))

	for _, tableName := range updateAuditTableFor {

		table := tableMap[tableName]
		auditTable := tableMap[tableName+"_audit"]
		existingColumns := auditTable.Columns

		existingColumnMap := make(map[string]api2go.ColumnInfo)
		for _, col := range existingColumns {
			existingColumnMap[col.Name] = col
		}

		tableColumnMap := make(map[string]api2go.ColumnInfo)
		for _, col := range table.Columns {
			tableColumnMap[col.Name] = col
		}

		newColsToAdd := make([]api2go.ColumnInfo, 0)

		for _, newCols := range table.Columns {

			_, ok := existingColumnMap[newCols.Name]
			if !ok {
				var newAuditCol api2go.ColumnInfo
				copier.Copy(&newAuditCol, &newCols)
				newColsToAdd = append(newColsToAdd, newAuditCol)
			}

		}

		if len(newColsToAdd) > 0 {

			for i := range config.Tables {

				if config.Tables[i].TableName == auditTable.TableName {
					config.Tables[i].Columns = append(config.Tables[i].Columns, newColsToAdd...)
				}
			}

		}

	}

	convertRelationsToColumns(newRelations, config)

}

func convertRelationsToColumns(relations []api2go.TableRelation, config *CmsConfig) {

	existingRelationMap := make(map[string]bool)

	for _, rel := range config.Relations {
		existingRelationMap[rel.Hash()] = true
	}

	for _, relation := range relations {

		if existingRelationMap[relation.Hash()] {
			log.Infof("Relation [%v] is already registered", relation.String())
			continue
		}
		log.Infof("Register relation [%v]", relation.String())
		//config.Relations = append(config.Relations, relation)
		config.AddRelations(relation)
		existingRelationMap[relation.Hash()] = true

		relation2 := relation.GetRelation()
		//log.Infof("Relation to table [%v]", relation.String())
		if relation2 == "belongs_to" || relation2 == "has_one" {
			fromTable := relation.Subject
			targetTable := relation.Object

			log.Infof("From table [%v] to table [%v]", fromTable, targetTable)
			isNullable := false
			if targetTable == "user" || targetTable == "usergroup" || relation2 == "has_one" {
				isNullable = true
			}

			col := api2go.ColumnInfo{
				Name:         relation.GetObject(),
				ColumnName:   relation.GetObjectName(),
				IsForeignKey: true,
				ColumnType:   "alias",
				IsNullable:   isNullable,
				ForeignKeyData: api2go.ForeignKeyData{
					TableName:  targetTable,
					ColumnName: "id",
					DataSource: "self",
				},
				DataType: "int(11)",
			}

			noMatch := true

			// there are going to be 2 tables sometimes which will be marked as "not top tables", so we cannot break after first match
			for i, t := range config.Tables {
				if t.TableName == fromTable {
					noMatch = false
					c := t.Columns

					exists := false
					for _, c1 := range c {
						if c1.ColumnName == col.ColumnName {
							exists = true
							break
						}
					}

					if !exists {
						c = append(c, col)
						config.Tables[i].Columns = c
					}

					log.Infof("Add column [%v] to table [%v]", col.ColumnName, t.TableName)
					if targetTable != "user" && relation.GetRelation() == "belongs_to" {
						config.Tables[i].IsTopLevel = false
						log.Infof("Table [%v] is not top level == %v", t.TableName, targetTable)
					}
				}

			}
			if noMatch {
				newTable := TableInfo{
					TableName: fromTable,
					Columns:   []api2go.ColumnInfo{col},
				}
				config.Tables = append(config.Tables, newTable)
				log.Infof("No matching table found: %v", relation)
				log.Infof("Created new table: %v", newTable.TableName)
			}
		} else if relation2 == "has_many" {

			fromTable := relation.GetSubject()
			targetTable := relation.GetObject()

			newTable := TableInfo{
				TableName:   relation.GetJoinTableName(),
				Columns:     make([]api2go.ColumnInfo, 0),
				IsJoinTable: true,
				IsTopLevel:  false,
			}

			col1 := api2go.ColumnInfo{
				Name:         fromTable + "_id",
				ColumnName:   relation.GetSubjectName(),
				ColumnType:   "alias",
				IsForeignKey: true,
				ForeignKeyData: api2go.ForeignKeyData{
					DataSource: "self",
					TableName:  fromTable,
					ColumnName: "id",
				},
				DataType: "int(11)",
			}

			newTable.Columns = append(newTable.Columns, col1)

			col2 := api2go.ColumnInfo{
				Name:         targetTable + "_id",
				ColumnName:   relation.GetObjectName(),
				ColumnType:   "alias",
				IsForeignKey: true,
				ForeignKeyData: api2go.ForeignKeyData{
					TableName:  targetTable,
					DataSource: "self",
					ColumnName: "id",
				},
				DataType: "int(11)",
			}

			newTable.Columns = append(newTable.Columns, col2)
			newTable.AddRelation(relation)
			//newTable.Relations = append(newTable.Relations, relation)
			log.Infof("Add column [%v] to table [%v]", col1.ColumnName, newTable.TableName)
			log.Infof("Add column [%v] to table [%v]", col2.ColumnName, newTable.TableName)

			config.Tables = append(config.Tables, newTable)

			if targetTable != "usergroup" {
				stateTable := TableInfo{
					TableName: newTable.TableName + "_state",
					Columns: []api2go.ColumnInfo{
						{
							ColumnName: "state",
							Name:       "state",
							ColumnType: "label",
							DataType:   "varchar(100)",
							IsNullable: false,
						},
						{
							ColumnName:   "smd_id",
							Name:         "smd_id",
							ColumnType:   "alias",
							DataType:     "int(11)",
							IsForeignKey: true,
							IsNullable:   false,
							ForeignKeyData: api2go.ForeignKeyData{
								DataSource: "self",
								TableName:  "smd",
								ColumnName: "id",
							},
						},
						{
							ColumnName:   newTable.TableName + "_id",
							Name:         newTable.TableName + "_id",
							ColumnType:   "alias",
							DataType:     "int(11)",
							IsForeignKey: true,
							IsNullable:   false,
							ForeignKeyData: api2go.ForeignKeyData{
								DataSource: "self",
								TableName:  newTable.TableName,
								ColumnName: "id",
							},
						},
					},
				}
				config.Tables = append(config.Tables, stateTable)
			}

		} else if relation2 == "has_many_and_belongs_to_many" {

			fromTable := relation.GetSubject()
			targetTable := relation.GetObject()

			newTable := TableInfo{
				TableName: relation.GetSubjectName() + "_" + relation.GetObjectName(),
				Columns:   make([]api2go.ColumnInfo, 0),
			}

			col1 := api2go.ColumnInfo{
				Name:         relation.GetSubjectName(),
				ColumnName:   relation.GetSubjectName(),
				IsForeignKey: true,
				ColumnType:   "alias",
				ForeignKeyData: api2go.ForeignKeyData{
					TableName:  fromTable,
					DataSource: "self",
					ColumnName: "id",
				},
				DataType: "int(11)",
			}

			newTable.Columns = append(newTable.Columns, col1)

			col2 := api2go.ColumnInfo{
				Name:         relation.GetObject(),
				ColumnName:   relation.GetObjectName(),
				ColumnType:   "alias",
				IsForeignKey: true,
				ForeignKeyData: api2go.ForeignKeyData{
					TableName:  targetTable,
					ColumnName: "id",
					DataSource: "self",
				},
				DataType: "int(11)",
			}

			newTable.Columns = append(newTable.Columns, col2)
			newTable.AddRelation(relation)
			//newTable.Relations = append(newTable.Relations, relation)
			log.Infof("Add column [%v] to table [%v]", col1.ColumnName, newTable.TableName)
			log.Infof("Add column [%v] to table [%v]", col2.ColumnName, newTable.TableName)

			config.Tables = append(config.Tables, newTable)

			if targetTable != "usergroup" {

				stateTable := TableInfo{
					TableName: newTable.TableName + "_state",
					Columns: []api2go.ColumnInfo{
						{
							ColumnName: "state",
							Name:       "state",
							ColumnType: "label",
							DataType:   "varchar(100)",
							IsNullable: false,
						},
						{
							ColumnName:   "smd_id",
							Name:         "smd_id",
							ColumnType:   "alias",
							IsForeignKey: true,
							DataType:     "int(11)",
							IsNullable:   false,
							ForeignKeyData: api2go.ForeignKeyData{
								TableName:  "smd",
								ColumnName: "id",
								DataSource: "self",
							},
						},
						{
							ColumnName:   newTable.TableName + "_id",
							Name:         newTable.TableName + "_id",
							ColumnType:   "alias",
							DataType:     "int(11)",
							IsForeignKey: true,
							IsNullable:   false,
							ForeignKeyData: api2go.ForeignKeyData{
								TableName:  newTable.TableName,
								ColumnName: "id",
								DataSource: "self",
							},
						},
					},
				}
				config.Tables = append(config.Tables, stateTable)
			}
		} else {
			log.Errorf("Failed to identify relation type: %v", relation)
		}

	}

}

func alterTableAddColumn(tableName string, colInfo *api2go.ColumnInfo, sqlDriverName string) string {
	sq := fmt.Sprintf("alter table %v add column %v", tableName, getColumnLine(colInfo, sqlDriverName))

	return sq
}

func CreateTable(tableInfo *TableInfo, db *sqlx.DB) {

	createTableQuery := MakeCreateTableQuery(tableInfo, db.DriverName())

	log.Infof("Create table query")
	log.Printf(createTableQuery)
	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Errorf("Failed to create table: %v", err)
	}
}

func MakeCreateTableQuery(tableInfo *TableInfo, sqlDriverName string) string {
	createTableQuery := fmt.Sprintf("create table %s (\n", tableInfo.TableName)

	columnStrings := []string{}
	colsDone := map[string]bool{}
	for _, c := range tableInfo.Columns {

		if c.ColumnName == "" && c.Name == "" {
			log.Errorf("Column name is null: %v", c)
		}

		if c.ColumnName == "" {
			c.ColumnName = c.Name
		}

		if strings.TrimSpace(c.ColumnName) == "" {
			continue
		}

		if colsDone[c.ColumnName] {
			continue
		}

		columnLine := getColumnLine(&c, sqlDriverName)

		colsDone[c.ColumnName] = true
		columnStrings = append(columnStrings, columnLine)
	}
	columnString := strings.Join(columnStrings, ",\n  ")
	createTableQuery += columnString + ")"
	return createTableQuery
}

func getColumnLine(c *api2go.ColumnInfo, sqlDriverName string) string {

	datatype := c.DataType

	if datatype == "" {
		datatype = "varchar(50)"
	}

	columnParams := []string{c.ColumnName, datatype}

	if datatype == "timestamp" && c.DefaultValue == "" {
		c.IsNullable = true
	}

	if !c.IsNullable {
		columnParams = append(columnParams, "not null")
	} else {
		columnParams = append(columnParams, "null")
	}

	if c.IsAutoIncrement {
		if sqlDriverName == "sqlite3" {
			columnParams = append(columnParams, "PRIMARY KEY")
		} else {
			columnParams = append(columnParams, "AUTO_INCREMENT PRIMARY KEY")
		}
	} else if c.IsPrimaryKey {
		columnParams = append(columnParams, "PRIMARY KEY")
	}

	if c.DefaultValue != "" {
		columnParams = append(columnParams, "default "+c.DefaultValue)
	}

	columnLine := strings.Join(columnParams, " ")
	return columnLine
}
