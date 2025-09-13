package repositories

import (
	// "fmt"
	// "github.com/goccy/go-yaml"
	"time"
)

// BaseModel contains common fields for all models.
type BaseModel struct {
	ID        string                 `json:"id" db:"primarykey"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
}

// // User represents a user entity.
// type User struct {
// 	BaseModel
// 	Email          string `json:"email" db:"unique,notnull"`
// 	Name           string `json:"name"`
// 	HashedPassword string `json:"hashed_password" db:"notnull"`
// }

// func UnmarshalWithBaseModel[T any](yamlData []byte, target *[]T) ([]T, error) {
// 	var apps []T
// 	var bases []BaseModel

// 	err := yaml.Unmarshal(yamlData, &apps)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = yaml.Unmarshal(yamlData, &bases)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(apps) != len(bases) {
// 		return nil, fmt.Errorf("mismatched lengths: %d apps, %d bases", len(apps), len(bases))
// 	}

// 	for i := range apps {
// 		if base, ok := any(apps[i]).(BaseModel); ok {
// 			base.ID = bases[i].ID
// 			base.CreatedAt = bases[i].CreatedAt
// 			base.Metadata = bases[i].Metadata
// 			base.Tags = bases[i].Tags
// 			apps[i] = any(base).(T)
// 		}
// 	}
// 	return apps, nil
// }
