package spider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"MyMovie/controllers/base"
	"MyMovie/models"
	// "MyMovie/modules/myspider"
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/pipeline"
	"github.com/hu17889/go_spider/core/spider"
)

const (
	doubanHot = "http://movie.douban.com/j/search_subjects?type=%s&tag=%s&sort=recommend&page_limit=%d&page_start=%d"

	pageLimit = 10000
	pageStart = 0
	pageEnd   = 10000
	videoType = "movie"
)

var (
	doubanTypes = map[string][]string{
		"tv":    {"热门", "美剧", "英剧", "韩剧", "日剧", "国产剧", "港剧", "日本动画"},
		"movie": {"热门", "最新", "经典", "可播放", "豆瓣高分", "冷门佳片", "华语", "欧美", "韩国", "日本", "动作", "喜剧", "爱情", "科幻", "悬疑", "恐怖", "成长"},
	}
	doubanDict = make(map[string]string)
)

type DoubanProcesser struct {
}

func insertMovie(movie *models.Movie) {
	o := orm.NewOrm()
	if movie.MovieTitle == "" {
		//insertSynchError(movie.MovieTitle, "Insert Movie error:Title is null", movie.XuandyId, o)
		return
	}
	var temp models.Movie
	err := o.QueryTable("Movie").Filter("DoubanId", movie.DoubanId).One(&temp)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		//insertSynchError(movie.MovieTitle, "Returned Multi Rows Not One:"+err.Error(), movie.XuandyId, o)
		return
	} else if err == orm.ErrNoRows {
		// 没有找到记录
		_, err = o.Insert(movie)
	} else {
		movie.MovieId = temp.MovieId
		if !strings.Contains(temp.Tag, movie.Tag) {
			movie.Tag = temp.Tag + "," + movie.Tag
		} else {
			movie.Tag = temp.Tag
		}
		_, err = o.Update(movie, "MovieType", "Rate", "IsBeetleSubject", "Playable", "Cover", "Cover_X", "Cover_Y", "IsNew", "Tag", "Douban_Url")
	}
}

func (this *DoubanProcesser) GetBaseInfo(p *page.Page) {
	urlTag := p.GetUrlTag()
	query := p.GetJson()
	query = query.GetPath("subjects")
	array, _ := query.Array()
	for _, obj := range array {
		movie := models.Movie{}
		m := obj.(map[string]interface{})
		id, err := strconv.ParseInt(m["id"].(string), 10, 32)
		if err == nil {
			movie.DoubanId = int(id)
		}
		movie.MovieTitle = m["title"].(string)
		rate, err := strconv.ParseFloat(m["rate"].(string), 32)
		if err == nil {
			movie.Rate = float32(rate)
		}
		movie.IsBeetleSubject = m["is_beetle_subject"].(bool)
		movie.Douban_Url = m["url"].(string)
		movie.Playable = m["playable"].(bool)
		movie.Cover = m["cover"].(string)

		cover_x := m["cover_x"].(json.Number)
		x, err := cover_x.Int64()
		if err == nil {
			movie.Cover_X = int(x)
		}

		cover_x = m["cover_y"].(json.Number)
		x, err = cover_x.Int64()
		if err == nil {
			movie.Cover_Y = int(x)
		}
		movie.IsNew = m["is_new"].(bool)
		movie.Tag = urlTag

		movie.MovieType = p.GetRequest().GetPostdata()
		insertMovie(&movie)
	}
}

func setValue(s *goquery.Selection, a *reflect.Value) {
	key := s.Find(".pl").Text()
	val, isExists := doubanDict[key]
	if isExists {
		a.FieldByName(val).SetString(s.Find(".attrs").Text())
	}
}

func (this *DoubanProcesser) GetDetail(p *page.Page) {
	reqUrl := regNumber.FindString(p.GetRequest().GetUrl())
	doubanId, _ := strconv.ParseInt(reqUrl, 10, 64)
	o := orm.NewOrm()

	movie := models.Movie{}
	err := o.QueryTable("Movie").Filter("DoubanId", int(doubanId)).One(&movie)

	if err == orm.ErrMultiRows {
		return
	} else if err == orm.ErrNoRows {
		fmt.Println("Error No Rows")
	}
	a := reflect.ValueOf(&movie).Elem()
	query := p.GetHtmlParser()
	info := query.Find(`div[id="content"] div[id="info"]`)
	info.ChildrenFiltered("span").Each(func(i int, s *goquery.Selection) {
		setValue(s, &a)
	})

	genre := ""
	info.Find(`span[property="v:genre"]`).Each(func(i int, sel *goquery.Selection) {
		genre += "/" + sel.Text()
	})
	genre = strings.Replace(genre, "/", "", 1)
	movie.MovieStyle = genre
	movie.Starring = info.Find(`span.actor span.attrs`).Text()
	movie.ReleaseTime = info.Find(`span[property="v:initialReleaseDate"]`).Text()
	movie.RunningTime = info.Find(`span[property="v:runtime"]`).Text()
	movie.MovieSummary = query.Find(`div[id="link-report"] span`).First().Text()
	movie.IsSynchDouban = true
	o.Update(&movie)
}

func (this *DoubanProcesser) Process(p *page.Page) {
	if !p.IsSucc() {
		fmt.Println("error:", p.Errormsg())
	} else {
		urlTag := p.GetUrlTag()
		if urlTag == "Get Detail" {
			this.GetDetail(p)
		} else {
			this.GetBaseInfo(p)
		}

	}
}

func NewDoubanProcesser() *DoubanProcesser {
	return &DoubanProcesser{}
}

type DoubanController struct {
	base.BaseController
}

func (this *DoubanController) SpiderSubjects() {
	sp := spider.NewSpider(NewDoubanProcesser(), "douban")
	// scheduler := myspider.NewRedisScheduler()
	// sp.SetScheduler(scheduler)
	sp.SetSleepTime("rand", 500, 2000)
	sp.SetThreadnum(threadnum)
	sp.AddPipeline(pipeline.NewPipelineConsole())

	header := make(http.Header)
	header.Set("Connection", "keep-alive")
	header.Set("Host", "movie.douban.com")
	header.Set("Referer", "http://movie.douban.com/")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36")

	for doubanType, tags := range doubanTypes {
		for _, tag := range tags {
			for i := pageStart; i < pageEnd; i = i + pageLimit {
				url := fmt.Sprintf(doubanHot, doubanType, tag, pageLimit, i)
				req := request.NewRequest(url, "json", tag, "GET", doubanType, header, nil, nil, nil)
				sp.AddRequest(req)
			}
		}
	}
	sp.Run()
	this.Ctx.WriteString("End Spider!")
}
func (this *DoubanController) SpiderMovieDetail() {
	sp := spider.NewSpider(NewDoubanProcesser(), "douban")
	sp.SetSleepTime("rand", 500, 3000)
	sp.SetThreadnum(threadnum)
	header := make(http.Header)
	header.Set("Connection", "keep-alive")
	header.Set("Host", "movie.douban.com")
	header.Set("Referer", "http://movie.douban.com/")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36")

	o := orm.NewOrm()
	var movies []models.Movie
	qs := o.QueryTable("Movie").Filter("IsSynchDouban", 0).Limit(-1)
	models.ListObjects(qs, &movies)
	for _, vlaue := range movies {
		req := request.NewRequest(vlaue.Douban_Url, "html", "Get Detail", "GET", "", header, nil, nil, nil)
		sp.AddRequest(req)
	}
	sp.SetThreadnum(threadnum)
	sp.Run()
	this.Ctx.WriteString("End Spider!")
}
func init() {
	doubanDict["导演"] = "Director"
	doubanDict["主演"] = "Starring"
	doubanDict["类型:"] = "MovieStyle"
	doubanDict["语言:"] = "DialogueLanguage"
	doubanDict["片长:"] = "RunningTime"
	doubanDict["又名:"] = "MovieTitleOther"
	doubanDict["制片国家/地区:"] = "ProductionArea"
	doubanDict["上映日期:"] = "ReleaseTime"
	doubanDict["IMDb链接:"] = "IMDB_Url"
}
