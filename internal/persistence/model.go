package persistence

type World struct {
	Name        string `gorm:"type:VARCHAR(50);primaryKey"`
	Game        string `gorm:"type:VARCHAR(50)"`
	Description string `gorm:"type:VARCHAR(1000)"`
}
