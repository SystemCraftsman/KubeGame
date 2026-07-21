package api

import (
	"fmt"
	"gorm.io/gorm"
	"systemcraftsman.com/kubegame/internal/persistence"
)

func validateAgainstBlueprint(db *gorm.DB, avatarName string, req *AvatarInstanceRequest) error {
	var avatarType persistence.AvatarType
	if result := db.Where("name = ?", avatarName).First(&avatarType); result.Error != nil {
		return fmt.Errorf("avatar type %q not found", avatarName)
	}

	if err := validateAttributes(db, avatarName, req.Attributes); err != nil {
		return err
	}
	if err := validateInventory(db, avatarName, req.Inventory); err != nil {
		return err
	}
	if err := validateAchievements(db, avatarName, req.Achievements); err != nil {
		return err
	}

	return nil
}

func validateAttributes(db *gorm.DB, avatarName string, attributes map[string]string) error {
	var validTypes []persistence.AttributeType
	db.Where("avatar_name = ?", avatarName).Find(&validTypes)

	validNames := make(map[string]bool)
	for _, t := range validTypes {
		validNames[t.Name] = true
	}

	for name := range attributes {
		if !validNames[name] {
			return fmt.Errorf("attribute %q is not defined in avatar type %q", name, avatarName)
		}
	}
	return nil
}

func validateInventory(db *gorm.DB, avatarName string, inventory []InventoryItem) error {
	var validTypes []persistence.InventoryType
	db.Where("avatar_name = ?", avatarName).Find(&validTypes)

	validCategories := make(map[string]bool)
	for _, t := range validTypes {
		validCategories[t.Category] = true
	}

	for _, item := range inventory {
		if !validCategories[item.Type] {
			return fmt.Errorf("inventory category %q is not defined in avatar type %q", item.Type, avatarName)
		}
	}
	return nil
}

func validateAchievements(db *gorm.DB, avatarName string, achievements []string) error {
	var validTypes []persistence.AchievementType
	db.Where("avatar_name = ?", avatarName).Find(&validTypes)

	validNames := make(map[string]bool)
	for _, t := range validTypes {
		validNames[t.Name] = true
	}

	for _, name := range achievements {
		if !validNames[name] {
			return fmt.Errorf("achievement %q is not defined in avatar type %q", name, avatarName)
		}
	}
	return nil
}
