package main

import (
	"fmt"

	"MyMovie/controllers/auth"
	"MyMovie/modules/utils"
	_ "MyMovie/routers"
	"MyMovie/setting"

	"github.com/astaxie/beego"
	//	"github.com/astaxie/beego/orm"
	"github.com/beego/social-auth"

	_ "github.com/go-sql-driver/mysql"
)

func initialize() {
	setting.LoadConfig()

	if err := utils.InitSphinxPools(); err != nil {
		beego.Error(fmt.Sprint("sphinx init pool", err))
	}

	setting.SocialAuth = social.NewSocial("/login/", auth.SocialAuther)
	setting.SocialAuth.ConnectSuccessURL = "/settings/profile"
	setting.SocialAuth.ConnectFailedURL = "/settings/profile"
	setting.SocialAuth.ConnectRegisterURL = "/register/connect"
	setting.SocialAuth.LoginURL = "/login"
}
func main() {
	beego.SetLogFuncCall(true)
	initialize()
	beego.Info("AppPath:", beego.AppConfigPath)
	if setting.IsProMode {
		beego.Info("Product mode enabled")
	} else {
		beego.Info("Develment mode enabled")
	}
	beego.Info(beego.BConfig.AppName, setting.APP_VER, setting.AppUrl)

	if !setting.IsProMode {
		beego.SetStaticPath("/static_source", "static_source")
		beego.BConfig.WebConfig.DirectoryIndex = true
	}
	beego.Run()
}

//func init() {
//	orm.RegisterDriver("mysql", orm.DRMySQL)
//	orm.RegisterDataBase("default", "mysql", "swtsoft:swtsoft@/MyMovie?charset=utf8")
//}
