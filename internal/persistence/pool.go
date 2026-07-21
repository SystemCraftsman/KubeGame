package persistence

import (
	"sync"

	"gorm.io/gorm"
)

var (
	dbPool sync.Map
)

func GetOrCreateConnection(hostName, username, password string) (*gorm.DB, error) {
	if val, ok := dbPool.Load(hostName); ok {
		db := val.(*gorm.DB)
		sqlDB, err := db.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				return db, nil
			}
		}
		dbPool.Delete(hostName)
	}

	db, err := CreateDatabaseConnection(hostName, username, password)
	if err != nil {
		return nil, err
	}

	actual, _ := dbPool.LoadOrStore(hostName, db)
	return actual.(*gorm.DB), nil
}

func CloseConnection(hostName string) {
	if val, ok := dbPool.LoadAndDelete(hostName); ok {
		sqlDB, err := val.(*gorm.DB).DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
