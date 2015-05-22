package main

import (
	"fmt"
	
	
	_ "MyMovie/routers"
	"MyMovie/setting"
	"MyMovie/modules/utils"
	"MyMovie/controllers/auth"
	
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
	beego.Info("AppPath:",beego.AppPath)
	if setting.IsProMode{
		beego.Info("Product mode enabled")
	}else{
		beego.Info("Develment mode enabled")
	}
	beego.Info(beego.AppName,setting.APP_VER,setting.AppUrl)
	
	if !setting.IsProMode{
		beego.SetStaticPath("/static_source","static_source")
		beego.DirectoryIndex=true
	}
	beego.Run()
}

//func init() {
//	orm.RegisterDriver("mysql", orm.DR_MySQL)
//	orm.RegisterDataBase("default", "mysql", "swtsoft:swtsoft@/MyMovie?charset=utf8")
//}
