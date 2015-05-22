package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type Movie struct {
	MovieId           int       `orm:"column(MovieId);pk"`
	MovieSummary      string    `orm:"column(MovieSummary)"`
	MovieImgSrc       string    `orm:"column(MovieImgSrc)"`
	MovieTitle        string    `orm:"column(MovieTitle)"`
	MovieTitleEN      string    `orm:"column(MovieTitle_EN)"`
	MovieTitleCN      string    `orm:"column(MovieTitle_CN)"`
	MovieTitleOther   string    `orm:"column(MovieTitle_Other)"`
	ProductionTime    string    `orm:"column(ProductionTime)"`
	ProductionCompany string    `orm:"column(ProductionCompany)"`
	DistributionFirm  string    `orm:"column(DistributionFirm)"`
	ProductionArea    string    `orm:"column(ProductionArea)"`
	ProductionCost    string    `orm:"column(ProductionCost)"`
	ShootingLocation  string    `orm:"column(ShootingLocation)"`
	ShootingDate      string    `orm:"column(ShootingDate)"`
	Director          string    `orm:"column(Director)"`
	ScreenWriter      string    `orm:"column(ScreenWriter)"`
	Producer          string    `orm:"column(Producer)"`
	MovieStyle        string    `orm:"column(MovieStyle)"`
	MovieType         string    `orm:"column(MovieType)"`
	Starring          string    `orm:"column(Starring)"`
	RunningTime       string    `orm:"column(RunningTime)"`
	ReleaseTime       string    `orm:"column(ReleaseTime)"`
	BoxOffice         string    `orm:"column(BoxOffice)"`
	DialogueLanguage  string    `orm:"column(DialogueLanguage)"`
	Synopsis          string    `orm:"column(Synopsis)"`
	Created           time.Time `orm:"auto_now_add;type(datetime)"`
	Updated           time.Time `orm:"auto_now;type(datetime)"`
	XuandyId          int       `orm:"column(XuandyId)"`
	IsSynchBaidu      bool      `orm:"column(IsSynchBaidu)"`
	DoubanId          int       `orm:"column(DoubanId)";json:"id"`
	Rate              float32   `orm:"column(Rate)";json:"rate"`
	IsBeetleSubject   bool      `orm:"column(Is_Beetle_Subject)";json:"is_beetle_subject"`
	Playable          bool      `orm:"column(Playable)";json:"playable"`
	Cover             string    `orm:"column(Cover)";json:"cover"`
	Cover_X           int       `orm:"column(Cover_X)";json:"cover_x"`
	Cover_Y           int       `orm:"column(Cover_Y)";json:"cover_y"`
	IsNew             bool      `orm:"column(Is_New)";json:"is_new"`
	Tag               string    `orm:"column(Tag)"`
	Douban_Url        string    `orm:"column(Douban_Url)";json:"url"`
	IMDB_Url          string    `orm:"column(IMDB_Url)"`
	IsSynchDouban     bool      `orm:"column(IsSynchDouban)"`
}

//type MovieType struct {
//}
type MovieDownloadUrl struct {
	UrlId       int    `orm:"column(UrlId);pk"`
	MovieId     int    `orm:"column(MovieId)"`
	DownloadUrl string `orm:"column(DownloadUrl)"`
	UrlTitle    string `orm:"column(UrlTitle)"`
}
type SynchError struct {
	SynchErrorId int       `orm:"column(SynchErrorId);pk"`
	MovieTitle   string    `orm:"column(MovieTitle)"`
	XuandyId     int       `orm:"column(XuandyId)"`
	ErrorMgs     string    `orm:"column(ErrorMgs)"`
	Created      time.Time `orm:"auto_now_add;type(datetime)"`
}

type Xuandy struct {
	Id        int       `orm:"column(Id);pk"`
	XuandyId  int       `orm:"column(XuandyId)"`
	Url       string    `orm:"column(Url)"`
	Created   time.Time `orm:"auto_now_add;type(datetime);column(Created)"`
	Updated   time.Time `orm:"auto_now;type(datetime);column(Updated)"`
	MovieType string    `orm:"column(MovieType)"`
}

func init() {
	orm.RegisterModel(new(Movie), new(MovieDownloadUrl), new(SynchError), new(Xuandy))
}
