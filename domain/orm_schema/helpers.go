package orm_schema

import "context"

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
