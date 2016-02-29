package controllers

import (
	"MyMovie/controllers/base"
	"MyMovie/models"
	"github.com/astaxie/beego/orm"
)

type MovieController struct {
	base.BaseController
}

//
func (this *MovieController) Category() {
	this.Data["IsHome"] = true
	this.TplName = "movie/category.html"
	slug := this.GetString(":type")
	o := orm.NewOrm()
	var (
		movies []models.Movie
	)
	pers := 10
	qs := o.QueryTable("Movie").Filter("MovieType", slug)
	cnt, _ := models.CountObjects(qs)
	pager := this.SetPaginator(pers, cnt)
	qs = qs.OrderBy("-Created").Limit(pers, pager.Offset())
	models.ListObjects(qs, &movies)
	this.Data["Movies"] = &movies
}

//
func (this *MovieController) Index() {
	this.Data["IsHome"] = true
	this.TplName = "index.html"
	o := orm.NewOrm()
	var (
		movies    []models.Movie
		movieList []models.Movie
		telList   []models.Movie
		videoList []models.Movie
	)
	qs := o.QueryTable("Movie").OrderBy("-Created")
	qsMovie := qs.Limit(10)
	models.ListObjects(qsMovie, &movies)
	this.Data["Movies"] = &movies

	qsMovie1 := qs.Filter("MovieType", "Movie").Limit(10)
	models.ListObjects(qsMovie1, &movieList)
	this.Data["MovieList"] = &movieList

	qsTel := qs.Filter("MovieType", "tv").Limit(10)
	models.ListObjects(qsTel, &telList)
	this.Data["TelList"] = &telList

	qsVideo := qs.Filter("MovieType", "Video").Limit(10)
	models.ListObjects(qsVideo, &videoList)
	this.Data["VideoList"] = &videoList
}

//
func (this *MovieController) GetMovieDetail() {
	id := this.Ctx.Input.Param(":id")

	this.TplName = "movie/movieDetail.html"
	var (
		movie models.Movie
		urls  []*models.MovieDownloadUrl
	)
	o := orm.NewOrm()

	o.QueryTable("Movie").Filter("MovieId", id).One(&movie)
	o.QueryTable("MovieDownloadUrl").Filter("MovieId", id).All(&urls)
	this.Data["Movie"] = &movie
	this.Data["Urls"] = &urls
}
