package orm_schema

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"strconv"
)

func GetAndStoreModels(persistent PersistentRepository, cache CacheRepository, ctx context.Context) error {
	models, err := persistent.GetAllModels()
	if err != nil {
		return err
	} else if models == nil {
		return nil
	}

	modelsMap := make(map[string]Model)
	for _, m := range *models {
		modelsMap[m.Name] = m
	}
	if err := cache.SetModels(&modelsMap, ctx); err != nil {
		return err
	}
	return nil
}

func GetAndStoreAssociations(persistent PersistentRepository, cache CacheRepository, ctx context.Context) error {
	associations, err := persistent.GetAllAssociations()
	if err != nil {
		return err
	} else if associations == nil {
		return nil
	}

	associationsMap := make(map[string]ModelAssociation)
	for _, m := range *associations {
		associationsMap[m.Name] = m
	}
	if err := cache.SetAssociations(&associationsMap, ctx); err != nil {
		return err
	}
	return nil
}

func (p *Projection) SanitizeValue(val any, exists bool) (any, error) {
	switch p.ModelType {
	case common.DatabaseTypeString:
		if exists && val != nil {
			return fmt.Sprint(val), nil
		} else if p.NotNull {
			return "", nil
		} else {
			return nil, nil
		}
	case common.DatabaseTypeBoolean:
		if exists && p.ModelType == p.SchemaType {
			return val, nil
		} else if exists {
			if val, err := strconv.ParseBool(fmt.Sprint(val)); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", p.ModelType, err)
			} else {
				return val, nil
			}
		} else {
			return false, nil
		}

	case common.DatabaseTypeNumber:
		if exists && val != nil && p.ModelType == p.SchemaType {
			return val, nil
		} else if exists && val != nil && p.SchemaType == common.DatabaseTypeString {
			if val, err := strconv.ParseFloat(fmt.Sprint(val), 64); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", p.ModelType, err)
			} else {
				return val, nil
			}
		} else if exists && val != nil && p.SchemaType == common.DatabaseTypeBoolean {
			if val, err := strconv.ParseBool(fmt.Sprint(val)); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", p.ModelType, err)
			} else {
				return val, nil
			}
		} else if p.NotNull {
			return 0, nil
		} else {
			return nil, nil
		}
	default:
		return val, nil
	}
}
