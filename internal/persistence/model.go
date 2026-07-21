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
}

type AvatarInstanceAchievement struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	AvatarInstanceID uint   `gorm:"index"`
	Name             string `gorm:"type:VARCHAR(100)"`
}
