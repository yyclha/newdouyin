package bootstrap

import (
	_ "douyin-backend/app/core/destroy"
	"douyin-backend/app/global/my_errors"
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/validator/common/register_validator"
	"douyin-backend/app/service/sys_log_hook"
	"douyin-backend/app/service/upload_file"
	"douyin-backend/app/utils/casbin_v2"
	"douyin-backend/app/utils/gorm_v2"
	"douyin-backend/app/utils/snow_flake"
	"douyin-backend/app/utils/validator_translation"
	"douyin-backend/app/utils/websocket/core"
	"douyin-backend/app/utils/yml_config"
	"douyin-backend/app/utils/zap_factory"
	"log"
	"os"
)

func checkRequiredFolders() {
	storageAppPath := variable.BasePath + "/storage/app"
	publicStoragePath := variable.BasePath + "/public/storage"

	if _, err := os.Stat(variable.BasePath + "/config/config.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/config/gorm_v2.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigGormNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/public/"); err != nil {
		log.Fatal(my_errors.ErrorsPublicNotExists + err.Error())
	}
	if _, err := os.Stat(variable.BasePath + "/storage/logs/"); err != nil {
		log.Fatal(my_errors.ErrorsStorageLogsNotExists + err.Error())
	}
	if err := os.MkdirAll(storageAppPath, os.ModePerm); err != nil {
		log.Fatal(my_errors.ErrorsSoftLinkCreateFail + err.Error())
	}

	if _, err := os.Lstat(publicStoragePath); err == nil {
		if err = os.RemoveAll(publicStoragePath); err != nil {
			log.Fatal(my_errors.ErrorsSoftLinkDeleteFail + err.Error())
		}
	} else if !os.IsNotExist(err) {
		log.Fatal(my_errors.ErrorsSoftLinkDeleteFail + err.Error())
	}

	if err := os.Symlink(storageAppPath, publicStoragePath); err != nil {
		log.Fatal(my_errors.ErrorsSoftLinkCreateFail + err.Error())
	}
}

func init() {
	// 1. Initialize project base path.

	// 2. Validate required folders and rebuild the public storage symlink.
	checkRequiredFolders()

	// 3. Register validators.
	register_validator.WebRegisterValidator()
	//register_validator.ApiRegisterValidator()

	// 4. Initialize config watchers.
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	// 5. Initialize global logger.
	variable.ZapLog = zap_factory.CreateZapFactory(sys_log_hook.ZapLogHandler)

	// 6. Initialize gorm clients according to config.
	if variable.ConfigGormv2Yml.GetInt("Gormv2.Mysql.IsInitGlobalGormMysql") == 1 {
		if dbMysql, err := gorm_v2.GetOneMysqlClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbMysql = dbMysql
		}
	}
	if variable.ConfigGormv2Yml.GetInt("Gormv2.Sqlserver.IsInitGlobalGormSqlserver") == 1 {
		if dbSqlserver, err := gorm_v2.GetOneSqlserverClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbSqlserver = dbSqlserver
		}
	}
	if variable.ConfigGormv2Yml.GetInt("Gormv2.PostgreSql.IsInitGlobalGormPostgreSql") == 1 {
		if dbPostgre, err := gorm_v2.GetOnePostgreSqlClient(); err != nil {
			log.Fatal(my_errors.ErrorsGormInitFail + err.Error())
		} else {
			variable.GormDbPostgreSql = dbPostgre
		}
	}

	// 6.1 Initialize async video upload workers after database clients are ready.
	upload_file.InitVideoUploadQueue()

	// 7. Initialize snowflake generator.
	variable.SnowFlake = snow_flake.CreateSnowflakeFactory()

	// 8. Start websocket hub if enabled.
	if variable.ConfigYml.GetInt("Websocket.Start") == 1 {
		variable.WebsocketHub = core.CreateHubFactory()
		if Wh, ok := variable.WebsocketHub.(*core.Hub); ok {
			go Wh.Run()
		}
	}

	// 9. Initialize casbin if enabled.
	if variable.ConfigYml.GetInt("Casbin.IsInit") == 1 {
		var err error
		if variable.Enforcer, err = casbin_v2.InitCasbinEnforcer(); err != nil {
			log.Fatal(err.Error())
		}
	}

	// 10. Register validator translations.
	if err := validator_translation.InitTrans("zh"); err != nil {
		log.Fatal(my_errors.ErrorsValidatorTransInitFail + err.Error())
	}
}
