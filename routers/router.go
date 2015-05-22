package routers

import (
	"MyMovie/controllers"
	"MyMovie/controllers/spider"
	"github.com/astaxie/beego"
)

func init() {
	spiderCtr := new(spider.SpiderController)
	beego.Router("/SpiderXuandyAll", spiderCtr, "get:SpiderXuandyAll")
	beego.Router("/SpiderXuandySynch", spiderCtr, "get:SynchXuandy")
	beego.Router("/SynchDownloadUrl", spiderCtr, "get:SynchDownloadUrl")

	baidkeCtrl := new(spider.BaiduBaikeController)
	beego.Router("/SpiderBaiduBaike", baidkeCtrl, "get:SpiderBaiduBaike")

	movie := new(controllers.MovieController)
	beego.Router("/", movie, "get:Index")
	beego.Router("/category/:type", movie, "get:Category")
	beego.Router("/movie/:id([0-9]+).html", movie, "get:GetMovieDetail")

	doubanCtrl := new(spider.DoubanController)
	beego.Router("/SpiderDoubanSubjects", doubanCtrl, "get:SpiderSubjects")
	beego.Router("/SpiderMovieDetail", doubanCtrl, "get:SpiderMovieDetail")
}
