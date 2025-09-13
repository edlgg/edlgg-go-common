package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo[T any] struct {
	db        *pgxpool.Pool
	tableName string
}

func NewPostgresRepo[T any](db *pgxpool.Pool, tableName string) (*PostgresRepo[T], error) {
	repo := &PostgresRepo[T]{db: db, tableName: tableName}

	// Validate or create the table
	err := repo.ensureTableExists()
	if err != nil {
		slog.Error("Failed to ensure table exists", "table", tableName, "error", err)
		return nil, fmt.Errorf("failed to ensure table exists: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepo[T]) ensureTableExists() error {
	// Check if the table exists
	tableExists, err := r.checkTableExists()
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if !tableExists {
		// Create an empty table
		err = r.createEmptyTable()
		if err != nil {
			return fmt.Errorf("failed to create empty table: %w", err)
		}
	}

	// Ensure all columns exist
	err = r.ensureColumnsExist()
	if err != nil {
		return fmt.Errorf("failed to ensure columns exist: %w", err)
	}

	return nil
}

func (r *PostgresRepo[T]) checkTableExists() (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables
            WHERE table_name = $1
        );
    `
	var exists bool
	err := r.db.QueryRow(context.Background(), query, r.tableName).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo[T]) createEmptyTable() error {
	query := fmt.Sprintf(`CREATE TABLE %s ();`, r.tableName)
	_, err := r.db.Exec(context.Background(), query)
	return err
}

func (r *PostgresRepo[T]) ensureColumnsExist() error {
	// Get the struct type of T
	var entity T
	t := reflect.TypeOf(entity)
	return r.processStructFields(t)
}

func (r *PostgresRepo[T]) processStructFields(t reflect.Type) error {
	// Iterate over the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		err := r.validateColumn(field)
		if err != nil {
			slog.Error("Failed to validate column", "field", field.Name, "error", err)
			return fmt.Errorf("failed to validate column %s: %w", field.Name, err)
		}
	}
	return nil
}

func (r *PostgresRepo[T]) validateColumn(field reflect.StructField) error {
	// Check if the field is an embedded struct (e.g., BaseModel)
	if field.Anonymous && field.Type.Kind() == reflect.Struct {
		// Recursively process the fields of the embedded struct
		return r.processStructFields(field.Type)
	}

	// Use the `json` tag as the column name
	columnName := field.Tag.Get("json")
	if columnName == "" {
		return fmt.Errorf("field %s has no json tag", field.Name)
	}

	// Get the SQL type for the field
	columnType, err := r.getSQLType(field.Type)
	if err != nil {
		return fmt.Errorf("unsupported type for field %s: %v", field.Name, err)
	}

	// Extract constraints from the `db` tag
	constraints := field.Tag.Get("db")

	// Check if the column exists
	columnExists, err := r.checkColumnExists(columnName)
	if err != nil {
		return err
	}

	// Add the column if it doesn't exist
	if !columnExists {
		err = r.addColumn(columnName, columnType, constraints)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepo[T]) checkColumnExists(columnName string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = $1 AND column_name = $2
        );
    `
	var exists bool
	err := r.db.QueryRow(context.Background(), query, r.tableName, columnName).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo[T]) addColumn(columnName, columnType string, constraints string) error {
	// Build the column definition with constraints
	columnDefinition := fmt.Sprintf("%s %s", columnName, columnType)

	if strings.Contains(constraints, "primarykey") {
		// Add the column as a primary key
		columnDefinition += " PRIMARY KEY"
	} else {
		if strings.Contains(constraints, "notnull") {
			columnDefinition += " NOT NULL"
		}
		if strings.Contains(constraints, "unique") {
			columnDefinition += " UNIQUE"
		}
	}

	// Construct the ALTER TABLE query
	query := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s;`, r.tableName, columnDefinition)
	_, err := r.db.Exec(context.Background(), query)
	return err
}

func (r *PostgresRepo[T]) getSQLType(goType reflect.Type) (string, error) {
	switch goType.Kind() {
	case reflect.String:
		return "TEXT", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER", nil
	case reflect.Float32, reflect.Float64:
		return "REAL", nil
	case reflect.Bool:
		return "BOOLEAN", nil
	case reflect.Struct:
		if goType == reflect.TypeOf(time.Time{}) {
			return "TIMESTAMP WITH TIME ZONE;", nil
		}
		return "", fmt.Errorf("unsupported Go type: struct")
	case reflect.Map:
		if goType.Key().Kind() == reflect.String && goType.Elem().Kind() == reflect.Interface {
			return "JSONB", nil
		}
		return "", fmt.Errorf("unsupported map type: %s", goType)
	case reflect.Slice:
		if goType.Elem().Kind() == reflect.String {
			return "TEXT[]", nil
		}
		return "", fmt.Errorf("unsupported slice type: %s", goType)
	default:
		return "", fmt.Errorf("unsupported Go type: %s", goType.Kind())
	}
}

func (r *PostgresRepo[T]) Set(entity *T) error {
	// Use reflection to get the type and value of the entity
	t := reflect.TypeOf(*entity)
	v := reflect.ValueOf(*entity)

	// Prepare slices for column names and placeholders
	var columns []string
	var placeholders []string
	var values []interface{}

	// Extract fields for the INSERT query
	err := r.extractFieldsForInsert(t, v, &columns, &placeholders, &values)
	if err != nil {
		return fmt.Errorf("failed to extract fields: %w", err)
	}

	// Build the INSERT query
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING;`,
		r.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute the query
	_, err = r.db.Exec(context.Background(), query, values...)
	if err != nil {
		slog.Error("Failed to insert entity", "table", r.tableName, "error", err)
		return fmt.Errorf("failed to insert entity into table %s: %w", r.tableName, err)
	}

	return nil
}

func (r *PostgresRepo[T]) extractFieldsForInsert(
	t reflect.Type,
	v reflect.Value,
	columns *[]string,
	placeholders *[]string,
	values *[]interface{},
) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Check if the field is an embedded struct
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recursively process the fields of the embedded struct
			err := r.extractFieldsForInsert(field.Type, value, columns, placeholders, values)
			if err != nil {
				return err
			}
			continue
		}

		// Use the `json` tag as the column name
		columnName := field.Tag.Get("json")
		if columnName == "" {
			continue // Skip fields without a `json` tag
		}

		// Add the column name and placeholder
		*columns = append(*columns, columnName)
		*placeholders = append(*placeholders, fmt.Sprintf("$%d", len(*values)+1))
		*values = append(*values, value.Interface())
	}
	return nil
}

func (r *PostgresRepo[T]) Delete(id string) error {
	// Build the DELETE query
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", r.tableName)

	// Execute the query
	_, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		slog.Error("Failed to delete entity", "table", r.tableName, "id", id, "error", err)
		return fmt.Errorf("failed to delete entity with id %s from table %s: %w", id, r.tableName, err)
	}

	return nil
}

func (r *PostgresRepo[T]) Query(
	whereClauses []WhereClause,
	sortBy string,
	descending bool,
	limit int,
	offset int,
) ([]T, error) {
	// Start building the SELECT query
	query := fmt.Sprintf("SELECT * FROM %s", r.tableName)
	var params []interface{}

	// Add WHERE clauses
	if len(whereClauses) > 0 {
		conditions := []string{}
		for i, clause := range whereClauses {
			if clause.Operator == "==" {
				clause.Operator = "=" // Normalize to SQL syntax
			}
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", clause.Field, clause.Operator, i+1))
			params = append(params, clause.Value)
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY clause
	if sortBy != "" {
		order := "ASC"
		if descending {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)
	}

	// Add LIMIT and OFFSET
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	// Execute the query
	rows, err := r.db.Query(context.Background(), query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse the results into a slice of T
	var results []T
	for rows.Next() {
		var entity T
		err := rows.Scan(r.getScanDestinations(&entity)...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, entity)
	}

	return results, nil
}

func (r *PostgresRepo[T]) getScanDestinations(entity *T) []interface{} {
	var destinations []interface{}
	r.collectScanDestinations(reflect.TypeOf(*entity), reflect.ValueOf(entity).Elem(), &destinations)
	return destinations
}

func (r *PostgresRepo[T]) collectScanDestinations(t reflect.Type, v reflect.Value, destinations *[]interface{}) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Check if the field is an embedded struct
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			r.collectScanDestinations(field.Type, value, destinations)
			continue
		}

		// Add the address of the field to the destinations
		*destinations = append(*destinations, value.Addr().Interface())
	}
}
