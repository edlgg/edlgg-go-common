package repositories

// This code has bugs and should be tested
// where filtering is not working
// where filtering is using or logic instead of and logic

// import (
// 	"fmt"
// 	"reflect"
// 	"sort"
// 	"sync"
// 	"log/slog"
// )

// // InMemoryRepo provides default implementations for repository methods.
// type InMemoryRepo[T any] struct {
// 	data      map[string]map[string]interface{}
// 	mu        *sync.RWMutex
// 	tableName string
// }

// // NewInMemoryDB initializes a new InMemoryRepository.
// func NewInMemoryRepo[T any](data map[string]map[string]interface{}, tableName string) *InMemoryRepo[T] {
// 	if _, exists := data[tableName]; !exists {
// 		data[tableName] = make(map[string]interface{})
// 	}
// 	return &InMemoryRepo[T]{
// 		data:      data,
// 		mu:        &sync.RWMutex{},
// 		tableName: tableName,
// 	}
// }

// func (r *InMemoryRepo[T]) Set(entity *T) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	// Use reflection to access the embedded BaseModel
// 	v := reflect.ValueOf(entity).Elem()
// 	baseModelField := v.FieldByName("BaseModel")
// 	if !baseModelField.IsValid() {
// 		return fmt.Errorf("entity does not implement BaseModel")
// 	}

// 	baseModel := baseModelField.Interface().(BaseModel)
// 	id := baseModel.ID
// 	if _, exists := r.data[r.tableName][id]; exists {
// 		return fmt.Errorf("entity with ID %s already exists", id)
// 	}
// 	r.data[r.tableName][id] = entity
// 	fmt.Printf("Created entity with ID %s\n", id)
// 	return nil
// }

// // Delete removes an entity by its ID.
// func (r *InMemoryRepo[T]) Delete(id string) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	if _, exists := r.data[r.tableName][id]; !exists {
// 		return fmt.Errorf("entity with ID %s not found", id)
// 	}
// 	delete(r.data[r.tableName], id)
// 	return nil
// }

// func (r *InMemoryRepo[T]) Query(
// 	whereClauses []WhereClause,
// 	sortBy string,
// 	descending bool,
// 	limit int,
// 	offset int,
// ) ([]T, error) {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()
// 	var results []T
// 	for _, entity := range r.data[r.tableName] {
// 		castedEntity, ok := entity.(*T)
// 		if !ok {
// 			return nil, fmt.Errorf("entity is not of type %T", new(T))
// 		}
// 		results = append(results, *castedEntity)
// 	}
// 	slog.Info("Querying in-memory repository", "table", r.tableName, "total", len(results) )
// 	if len(whereClauses) > 0 {
// 		results = filterByWhereClauses(results, whereClauses)
// 	}
// 	slog.Info("After filtering", "table", r.tableName, "total", len(results) )
// 	sortKey := "ID"
// 	if sortBy != "" {
// 		sortKey = sortBy
// 	}
// 	sort.Slice(results, func(i, j int) bool {
// 		v1 := reflect.ValueOf(results[i])
// 		v2 := reflect.ValueOf(results[j])

// 		// Dereference pointers if necessary
// 		if v1.Kind() == reflect.Ptr {
// 			v1 = v1.Elem()
// 		}
// 		if v2.Kind() == reflect.Ptr {
// 			v2 = v2.Elem()
// 		}

// 		// Get the fields by the sort key
// 		field1 := v1.FieldByName(sortKey)
// 		field2 := v2.FieldByName(sortKey)

// 		// Ensure the fields are valid
// 		if !field1.IsValid() || !field2.IsValid() {
// 			return false // Skip sorting if fields are invalid
// 		}

// 		// Handle sorting based on field types
// 		switch field1.Kind() {
// 		case reflect.String:
// 			if descending {
// 				return field1.String() > field2.String()
// 			}
// 			return field1.String() < field2.String()
// 		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 			if descending {
// 				return field1.Int() > field2.Int()
// 			}
// 			return field1.Int() < field2.Int()
// 		case reflect.Float32, reflect.Float64:
// 			if descending {
// 				return field1.Float() > field2.Float()
// 			}
// 			return field1.Float() < field2.Float()
// 		default:
// 			return false // Unsupported type, no sorting applied
// 		}
// 	})

// 	if offset != 0 {
// 		if offset >= len(results) {
// 			return nil, fmt.Errorf("offset %d is out of bounds", offset)
// 		}
// 		results = results[offset:]
// 	}
// 	if limit > 0 && limit < len(results) {
// 		results = results[:limit]
// 	}
// 	return results, nil
// }

// func filterByWhereClauses[T any](entities []T, whereClauses []WhereClause) []T {
// 	var filtered []T
// 	for _, entity := range entities {
// 		matches := false
// 		for _, clause := range whereClauses {
// 			v := reflect.ValueOf(entity)
// 			// Check if the value is a pointer and dereference it
// 			if v.Kind() == reflect.Ptr {
// 				v = v.Elem()
// 			}
// 			field := v.FieldByName(clause.Field)
// 			if !field.IsValid() {
// 				continue // Skip if field does not exist
// 			}
// 			slog.Info("Filtering", "field", clause.Field, "operator", clause.Operator, "value", clause.Value, "entityFieldValue", field.Interface())
// 			switch clause.Operator {
// 			case "==":
// 				if field.Interface() == clause.Value {
// 					matches = true
// 				}
// 			case "!=":
// 				if field.Interface() != clause.Value {
// 					matches = true
// 				}
// 			case ">":
// 				if field.Kind() == reflect.Int && field.Int() > int64(clause.Value.(int)) {
// 					matches = true
// 				}
// 			case "<":
// 				if field.Kind() == reflect.Int && field.Int() < int64(clause.Value.(int)) {
// 					matches = true
// 				}
// 			}
// 		}
// 		if matches {
// 			filtered = append(filtered, entity)
// 		}
// 	}
// 	return filtered
// }
