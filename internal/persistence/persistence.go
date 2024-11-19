package persistence

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/url"
	"os"
	"systemcraftsman.com/kubegame/internal/common"
)

func CreateDatabaseConnection(hostName string, username string, password string) (*gorm.DB, error) {
	databaseType := os.Getenv(common.EnvVarDatabaseType)
	switch databaseType {
	case "postgres":
		return createPostgresConnection(hostName, username, password)
	default:
		return nil, fmt.Errorf("unsupported database type : %s", databaseType)
	}
}

func createPostgresConnection(hostName string, username string, password string) (*gorm.DB, error) {
	var databaseUrl string
	if os.Getenv(common.EnvVarAppEnv) == "development" {
		databaseUrl = fmt.Sprintf("postgresql://%s/%s",
			"localhost:"+os.Getenv(common.EnvVarDatabasePortDevelopment),
			os.Getenv(common.EnvVarDatabaseName))
	} else {
		databaseUrl = fmt.Sprintf("postgresql://%s/%s",
			hostName+":"+os.Getenv(common.EnvVarDatabasePort),
			os.Getenv(common.EnvVarDatabaseName))
	}

	databaseDsn, err := url.Parse(databaseUrl)
	if err != nil {
		return nil, err
	}

	databaseDsn.User = url.UserPassword(username, password)

	return gorm.Open(postgres.Open(databaseDsn.String()), &gorm.Config{})
}
