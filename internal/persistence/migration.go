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
		&AvatarInstance{},
		&AvatarInstanceAttribute{},
		&AvatarInstanceInventoryItem{},
		&AvatarInstanceAchievement{},
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
	models = append(models, AreaModels()...)
	return models
}
