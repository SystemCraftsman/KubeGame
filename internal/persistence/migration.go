package persistence

import "gorm.io/gorm"

func RunMigrations(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}

func WorldModels() []interface{} {
	return []interface{}{
		&World{},
	}
}

func AvatarModels() []interface{} {
	return []interface{}{
		&AvatarType{},
		&AttributeType{},
		&InventoryType{},
		&AchievementType{},
		&CustomizationTypeRecord{},
		&CustomizationOption{},
		&AvatarInstance{},
		&AvatarInstanceAttribute{},
		&AvatarInstanceInventoryItem{},
		&AvatarInstanceAchievement{},
		&AvatarInstanceCustomization{},
	}
}

func ItemModels() []interface{} {
	return []interface{}{
		&ItemDefinition{},
		&ItemEffectRecord{},
		&ActivePowerup{},
	}
}

func AreaModels() []interface{} {
	return []interface{}{
		&Area{},
		&AreaConnection{},
		&AreaPropertyRecord{},
	}
}

func AllModels() []interface{} {
	var models []interface{}
	models = append(models, WorldModels()...)
	models = append(models, AvatarModels()...)
	models = append(models, ItemModels()...)
	models = append(models, AreaModels()...)
	return models
}
