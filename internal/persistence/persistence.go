package persistence

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/url"
	"os"
	"systemcraftsman.com/kubegame/internal/common"
)

func CreateDatabaseConnection(hostName string) (*gorm.DB, error) {
	databaseType := os.Getenv(common.EnvVarDatabaseType)
	switch databaseType {
	case "postgres":
		return createPostgresConnection(hostName)
	default:
		return nil, fmt.Errorf("unsupported database type : %s", databaseType)
	}
}

func createPostgresConnection(hostName string) (*gorm.DB, error) {
	databaseUrl := fmt.Sprintf("postgresql://%s/%s",
		hostName+":"+os.Getenv(common.EnvVarDatabasePort),
		os.Getenv(common.EnvVarDatabaseName))

	databaseDsn, err := url.Parse(databaseUrl)
	if err != nil {
		return nil, err
	}

	databaseDsn.User = url.UserPassword(
		os.Getenv(common.EnvVarDatabaseAdminUser),
		os.Getenv(common.EnvVarDatabaseAdminPassword))

	return gorm.Open(postgres.Open(databaseDsn.String()), &gorm.Config{})
}
