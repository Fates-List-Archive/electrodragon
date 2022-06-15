package utils

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var sqlxPool *sqlx.DB

type Schema struct {
	TableName  string  `json:"table_name"`
	ColumnName string  `json:"column_name"`
	Type       string  `json:"type"`
	IsNullable bool    `json:"nullable"`
	Array      bool    `json:"array"`
	DefaultSQL *string `json:"default_sql"`
	DefaultVal any     `json:"default_val"`
	Secret     bool    `json:"secret"`
}

func IsSecret(tableName, columnName string) bool {
	colArray := [2]string{tableName, columnName}

	secretCols := [][2]string{
		{
			"bots", "api_token",
		},
		{
			"bots", "webhook_secret",
		},
		{
			"users", "api_token",
		},
		{
			"users", "staff_password",
		},
		{
			"users", "totp_shared_key",
		},
		{
			"users", "supabase_id",
		},
		{
			"servers", "api_token",
		},
		{
			"servers", "webhook_secret",
		},
	}

	for _, col := range secretCols {
		if colArray == col {
			return true
		}
	}
	return false
}

func ConnectToDBIf() error {
	if sqlxPool == nil {
		db, err := sqlx.Connect("postgres", "sslmode=disable")
		if err != nil {
			return err
		}

		sqlxPool = db
	}
	return nil
}

type schemaData struct {
	ColumnDefault *string `db:"column_default"`
	TableSchema   string  `db:"table_schema"`
	TableName     string  `db:"table_name"`
	ColumnName    string  `db:"column_name"`
	DataType      string  `db:"data_type"`
	ElementType   *string `db:"element_type"`
	IsNullable    string  `db:"is_nullable"`
}

// Filter the postgres schema
type SchemaFilter struct {
	TableName string
}

func GetSchema(ctx context.Context, pool *pgxpool.Pool, opts SchemaFilter) ([]Schema, error) {
	var sqlString string = `
	SELECT c.is_nullable, c.table_name, c.column_name, c.column_default, c.data_type AS data_type, e.data_type AS element_type FROM information_schema.columns c LEFT JOIN information_schema.element_types e
	ON ((c.table_catalog, c.table_schema, c.table_name, 'TABLE', c.dtd_identifier)
= (e.object_catalog, e.object_schema, e.object_name, e.object_type, e.collection_type_identifier))
WHERE table_schema = 'public' order by table_name, ordinal_position
`
	if sqlxPool == nil {
		err := ConnectToDBIf()
		if err != nil {
			panic(err)
		}
	}

	rows, err := sqlxPool.Queryx(sqlString)

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	fmt.Println(cols)

	var result []Schema

	for rows.Next() {
		var schema Schema

		var data schemaData
		err = rows.StructScan(&data)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if opts.TableName != "" && opts.TableName != data.TableName {
			continue
		}

		// Create new transaction to get default column
		if data.ColumnDefault != nil && *data.ColumnDefault != "" {
			tx, err := pool.Begin(ctx)
			if err != nil {
				return nil, err
			}

			var defaultV any

			err = tx.QueryRow(ctx, "SELECT "+*data.ColumnDefault).Scan(&defaultV)

			if err != nil {
				return nil, err
			}

			fmt.Println(data.ColumnName, reflect.TypeOf(defaultV))

			err = tx.Rollback(ctx)

			if err != nil {
				return nil, err
			}

			// Check for [16]uint8 case
			if defaultVal, ok := defaultV.([16]uint8); ok {
				defaultV = fmt.Sprintf("%x-%x-%x-%x-%x", defaultVal[0:4], defaultVal[4:6], defaultVal[6:8], defaultVal[8:10], defaultVal[10:16])
			}

			schema.DefaultVal = defaultV
		} else {
			schema.DefaultVal = nil
		}

		// Now check if the column is tagged properly
		if _, err := sqlxPool.Queryx("SELECT _lynxtag FROM" + data.TableName); err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("Tagging", data.TableName)
				_, err := sqlxPool.Exec("ALTER TABLE " + data.TableName + " ADD COLUMN _lynxtag uuid not null unique default uuid_generate_v4()")
				if err != nil {
					return nil, err
				}
			}
		}

		schema.ColumnName = data.ColumnName
		schema.TableName = data.TableName
		schema.DefaultSQL = data.ColumnDefault

		schema.IsNullable = (data.IsNullable == "YES")

		if data.DataType == "ARRAY" {
			schema.Type = *data.ElementType
			schema.Array = true
		} else {
			schema.Type = data.DataType
		}

		schema.Secret = IsSecret(data.TableName, data.ColumnName)

		result = append(result, schema)
	}

	return result, nil
}
