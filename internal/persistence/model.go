package persistence

type World struct {
	Name        string `gorm:"type:VARCHAR(50);primaryKey"`
	Game        string `gorm:"type:VARCHAR(50)"`
	Description string `gorm:"type:VARCHAR(1000)"`
}

type AvatarType struct {
	Name        string `gorm:"type:VARCHAR(50);primaryKey"`
	Game        string `gorm:"type:VARCHAR(50)"`
	Type        string `gorm:"type:VARCHAR(100)"`
	Description string `gorm:"type:VARCHAR(1000)"`
}

type AttributeType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	AvatarName string `gorm:"type:VARCHAR(50)"`
	Name       string `gorm:"type:VARCHAR(100)"`
	ValueType  string `gorm:"type:VARCHAR(50)"`
}

type InventoryType struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	AvatarName string `gorm:"type:VARCHAR(50)"`
	Name       string `gorm:"type:VARCHAR(100)"`
	Category   string `gorm:"type:VARCHAR(100)"`
}

type AchievementType struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	AvatarName  string `gorm:"type:VARCHAR(50)"`
	Name        string `gorm:"type:VARCHAR(100)"`
	Description string `gorm:"type:VARCHAR(1000)"`
}

type AvatarInstance struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"type:VARCHAR(100);uniqueIndex"`
	AvatarName string `gorm:"type:VARCHAR(50)"`
	Game       string `gorm:"type:VARCHAR(50)"`
}

type AvatarInstanceAttribute struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	Name             string `gorm:"type:VARCHAR(100)"`
	Value            string `gorm:"type:VARCHAR(500)"`
}

type AvatarInstanceInventoryItem struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	Name             string `gorm:"type:VARCHAR(100)"`
	Type             string `gorm:"type:VARCHAR(100)"`
	Quantity         int
	Equipped         bool `gorm:"default:false"`
}

type AvatarInstanceAchievement struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	Name             string `gorm:"type:VARCHAR(100)"`
}

type CustomizationTypeRecord struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	AvatarName string `gorm:"type:VARCHAR(50);index"`
	Name       string `gorm:"type:VARCHAR(100)"`
}

type CustomizationOption struct {
	ID                uint   `gorm:"primaryKey;autoIncrement"`
	CustomizationName string `gorm:"type:VARCHAR(100);index"`
	AvatarName        string `gorm:"type:VARCHAR(50);index"`
	Value             string `gorm:"type:VARCHAR(100)"`
}

type AvatarInstanceCustomization struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	Name             string `gorm:"type:VARCHAR(100)"`
	Value            string `gorm:"type:VARCHAR(100)"`
}

type ItemDefinition struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:VARCHAR(100);uniqueIndex:idx_item_game"`
	Game      string `gorm:"type:VARCHAR(50);uniqueIndex:idx_item_game"`
	Category  string `gorm:"type:VARCHAR(50)"`
	Rarity    string `gorm:"type:VARCHAR(50)"`
	Stackable bool
	MaxStack  int
	Duration  int
}

type ItemEffectRecord struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	ItemName  string `gorm:"type:VARCHAR(100);index:idx_effect_item_game"`
	Game      string `gorm:"type:VARCHAR(50);index:idx_effect_item_game"`
	Attribute string `gorm:"type:VARCHAR(100)"`
	Modifier  string `gorm:"type:VARCHAR(50)"`
}

type ActivePowerup struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	ItemName         string `gorm:"type:VARCHAR(100)"`
	ActivatedAt      int64
	ExpiresAt        int64
}

type Area struct {
	Name        string `gorm:"type:VARCHAR(50);primaryKey"`
	Game        string `gorm:"type:VARCHAR(50)"`
	World       string `gorm:"type:VARCHAR(50)"`
	Description string `gorm:"type:VARCHAR(1000)"`
}

type AreaConnection struct {
	ID         uint   `gorm:"primaryKey;autoIncrement"`
	AreaName   string `gorm:"type:VARCHAR(50);index"`
	ConnectsTo string `gorm:"type:VARCHAR(50)"`
}

type AreaPropertyRecord struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	AreaName string `gorm:"type:VARCHAR(50);index"`
	Name     string `gorm:"type:VARCHAR(100)"`
	Value    string `gorm:"type:VARCHAR(500)"`
}
