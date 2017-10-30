// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/notifier"
	"github.com/vmware/harbor/src/common/scheduler"
	"github.com/vmware/harbor/src/ui/api"
	_ "github.com/vmware/harbor/src/ui/auth/db"
	_ "github.com/vmware/harbor/src/ui/auth/ldap"
	_ "github.com/vmware/harbor/src/ui/auth/uaa"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/filter"
	"github.com/vmware/harbor/src/ui/proxy"
	"github.com/vmware/harbor/src/ui/service/token"
)

const (
	adminUserID = 1
)

func updateInitPassword(userID int, password string) error {
	queryUser := models.User{UserID: userID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		return fmt.Errorf("Failed to get user, userID: %d %v", userID, err)
	}
	if user == nil {
		return fmt.Errorf("user id: %d does not exist", userID)
	}
	if user.Salt == "" {
		salt := utils.GenerateRandomString()

		user.Salt = salt
		user.Password = password
		err = dao.ChangeUserPassword(*user)
		if err != nil {
			return fmt.Errorf("Failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}

		log.Infof("User id: %d updated its encypted password successfully.", userID)
	} else {
		log.Infof("User id: %d already has its encrypted password.", userID)
	}
	return nil
}

func main() {
	beego.BConfig.WebConfig.Session.SessionOn = true
	//TODO
	redisURL := os.Getenv("_REDIS_URL")
	if len(redisURL) > 0 {
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = redisURL
	}
	beego.AddTemplateExt("htm")

	log.Info("initializing configurations...")
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	log.Info("configurations initialization completed")
	token.InitCreators()
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}
	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if config.WithClair() {
		clairDBPassword, err := config.ClairDBPassword()
		if err != nil {
			log.Fatalf("failed to load clair database information: %v", err)
		}
		if err := dao.InitClairDB(clairDBPassword); err != nil {
			log.Fatalf("failed to initialize clair database: %v", err)
		}
	}

	password, err := config.InitialAdminPassword()
	if err != nil {
		log.Fatalf("failed to get admin's initia password: %v", err)
	}
	if err := updateInitPassword(adminUserID, password); err != nil {
		log.Error(err)
	}

	//Enable the policy scheduler here.
	scheduler.DefaultScheduler.Start()

	//Subscribe the policy change topic.
	notifier.Subscribe(notifier.ScanAllPolicyTopic, &notifier.ScanPolicyNotificationHandler{})

	//Get policy configuration.
	scanAllPolicy := config.ScanAllPolicy()
	if scanAllPolicy.Type == notifier.PolicyTypeDaily {
		dailyTime := 0
		if t, ok := scanAllPolicy.Parm["daily_time"]; ok {
			if reflect.TypeOf(t).Kind() == reflect.Int {
				dailyTime = t.(int)
			}
		}

		//Send notification to handle first policy change.
		notifier.Publish(notifier.ScanAllPolicyTopic, notifier.ScanPolicyNotification{Type: scanAllPolicy.Type, DailyTime: (int64)(dailyTime)})
	}

	filter.Init()
	beego.InsertFilter("/*", beego.BeforeRouter, filter.SecurityFilter)

	initRouters()
	if err := api.SyncRegistry(config.GlobalProjectMgr); err != nil {
		log.Error(err)
	}
	log.Info("Init proxy")
	proxy.Init()
	//go proxy.StartProxy()
	beego.Run()
}
