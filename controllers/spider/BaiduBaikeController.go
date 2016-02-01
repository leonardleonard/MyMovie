//根据名称爬取百度百科信息
//

package spider

import (
	"MyMovie/controllers/base"
	"MyMovie/models"
	"MyMovie/modules/myspider"
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/common/request"
	"github.com/hu17889/go_spider/core/spider"

	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	baiduBaikeUrl = "http://baike.baidu.com/search/word?word=%s"
)

//
type BaiduBaikeProcesser struct {
}

//
func NewBaiduBaikeProcesser() *BaiduBaikeProcesser {
	return &BaiduBaikeProcesser{}
}

//
func (this *BaiduBaikeProcesser) getDetail1(p *page.Page, a *reflect.Value, movie *models.Movie) bool {

	query := p.GetHtmlParser()
	lemmaContent := query.Find(`div[id="lemmaContent-0"]`)
	summary := ""
	lemmaContent.ChildrenFiltered(".para").Each(func(i int, s *goquery.Selection) {
		summary += s.Text() + "\r\n"
	})
	movie.MovieSummary = summary

	this.getBaseInfo(lemmaContent, a)
	moviePlot := ""
	lemmaContent.Find("div.movie-plot").ChildrenFiltered(".para").Each(func(i int, s *goquery.Selection) {
		moviePlot += s.Text()
	})
	movie.Synopsis = moviePlot

	title := query.Find("div.lemmaTitleH1 span").First().Text()
	return isValid(title, *movie)
}

//
func (this *BaiduBaikeProcesser) Process(p *page.Page) {
	if !p.IsSucc() {
		PrintInfo(p.Errormsg())
		return
	}
	urlTag := p.GetUrlTag()
	id, _ := strconv.ParseInt(urlTag, 10, 64)

	var movie models.Movie
	movie.MovieId = int(id)
	o := orm.NewOrm()
	if err := o.Read(&movie); err != nil {
		PrintInfo(err)
		return
	}
	a := reflect.ValueOf(&movie).Elem()

	if !this.getDetail1(p, &a, &movie) {
		if !this.getDetail2(p, &a, &movie) {
			if !this.getTelDetail(p, &a, &movie) {
				return
			}
		}
	}

	movie.IsSynchBaidu = true
	if num, err := o.Update(&movie); err == nil {
		PrintInfo(num)
	} else {
		PrintInfo(err)
	}
}

//
func (this *BaiduBaikeProcesser) getBaseInfo(lemmaContent *goquery.Selection, a *reflect.Value) {
	baseInfoWrap := lemmaContent.Find(`div[id="baseInfoWrapDom"]`)
	leftItem := baseInfoWrap.Find(`div.baseInfoLeft`).Find("div.biItem")
	leftItem.Each(func(i int, s *goquery.Selection) {
		this.setValue(s, a)
	})
	rightItem := baseInfoWrap.Find(`div.baseInfoRight`).Find("div.biItem")
	rightItem.Each(func(i int, s *goquery.Selection) {
		this.setValue(s, a)
	})
}

//
func (this *BaiduBaikeProcesser) setValue(s *goquery.Selection, a *reflect.Value) {
	key := s.Find(".biItemInner .biTitle").Text()
	key = strings.Replace(key, "    ", "", -1)
	val, isExists := baikeDict[key]
	if isExists {
		a.FieldByName(val).SetString(s.Find(".biItemInner .biContent").Text())
	}
}

//
func isValid(title string, movie models.Movie) bool {
	if title == "" {
		PrintInfo("Title Is Null!")
		return false
	}
	if strings.HasPrefix(movie.MovieSummary, "UGC时期 百度百科历史首页图片") {
		PrintInfo("UGC时期 百度百科历史首页图片")
		return false
	}
	if title == movie.MovieTitle || movie.MovieTitle == movie.MovieTitleCN || movie.MovieTitle == movie.MovieTitleEN {
		return true
	}
	if strings.Contains(movie.MovieTitle, title) || strings.HasPrefix(title, movie.MovieTitle) || strings.Contains(movie.MovieTitleOther, movie.MovieTitle) || strings.HasPrefix(movie.MovieTitleCN, movie.MovieTitle) {
		return true
	}
	if strings.Replace(title, "·", "", -1) == strings.Replace(movie.MovieTitle, "·", "", -1) || strings.Replace(movie.MovieTitleCN, "·", "", -1) == strings.Replace(movie.MovieTitle, "·", "", -1) {
		return true
	}
	if strings.Replace(title, "：", "", -1) == strings.Replace(movie.MovieTitle, "：", "", -1) || strings.Replace(movie.MovieTitleCN, "：", "", -1) == strings.Replace(movie.MovieTitle, "：", "", -1) {
		return true
	}
	PrintInfo("其他原因！")
	return false
}

//
func (this *BaiduBaikeProcesser) getDetail2(p *page.Page, a *reflect.Value, movie *models.Movie) bool {

	query := p.GetHtmlParser()
	content := query.Find(`div[id="sec-content0"]`)

	container := query.Find(`div[id="card-container"]`)
	summary := ""
	container.Find(".para").Each(func(i int, s *goquery.Selection) {
		summary += s.Text() + "\r\n"
	})
	movie.MovieSummary = summary
	this.getBaseInfo(content, a)
	moviePlot := ""
	content.Find(`div[id="lemmaContent-0"]`).ChildrenFiltered(".para").Each(func(i int, s *goquery.Selection) {
		moviePlot += s.Text()
	})
	movie.Synopsis = moviePlot

	title := content.Find(`h1.maintitle span`).First().Text()
	return isValid(title, *movie)
}

//
func (this *BaiduBaikeProcesser) getTelDetail(p *page.Page, a *reflect.Value, movie *models.Movie) bool {

	query := p.GetHtmlParser()

	container := query.Find(`div[d="posterCon"] dd.desc`)
	summary := ""
	container.Find(".para").Each(func(i int, s *goquery.Selection) {
		summary += s.Text() + "\r\n"
	})
	movie.MovieSummary = summary

	content := query.Find(`div[id="content-wrap"] div[id="sec-content0"] div[id="lemmaContent-0"]`)
	this.getBaseInfo(content, a)
	moviePlot := ""
	content.ChildrenFiltered(".para").Each(func(i int, s *goquery.Selection) {
		moviePlot += s.Text()
	})
	movie.Synopsis = moviePlot

	title := query.Find(`div[d="posterCon"] dt.title h1`).First().Text()
	return isValid(title, *movie)
}

type BaiduBaikeController struct {
	base.BaseController
}

func (this *BaiduBaikeController) SpiderBaiduBaike() {
	sp := spider.NewSpider(NewBaiduBaikeProcesser(), "BaiduBaike")
	scheduler := myspider.NewRedisScheduler()
	sp.SetScheduler(scheduler)
	// req := request.NewRequest(fmt.Sprintf(baiduBaikeUrl, "记忆碎片"), "html", strconv.Itoa(242), "GET", "", nil, nil, nil, nil)
	// sp.AddRequest(req)
	o := orm.NewOrm()
	var movies []models.Movie
	qs := o.QueryTable("Movie").Filter("DoubanId", 0).Filter("IsSynchBaidu", 0).Limit(-1)
	models.ListObjects(qs, &movies)
	for _, vlaue := range movies {

		if vlaue.MovieTitle == "" {
			PrintInfo("Title Is Null!")
			continue
		}
		if strings.Contains(vlaue.MovieTitle, "/") {
			array := strings.Split(vlaue.MovieTitle, "/")
			for _, title := range array {
				req := request.NewRequest(fmt.Sprintf(baiduBaikeUrl, title), "html", strconv.Itoa(vlaue.MovieId), "GET", "", nil, nil, nil, nil)
				sp.AddRequest(req)
			}
		} else if strings.Contains(vlaue.MovieTitle, "：") {
			array := strings.Split(vlaue.MovieTitle, "：")
			for _, title := range array {
				req := request.NewRequest(fmt.Sprintf(baiduBaikeUrl, title), "html", strconv.Itoa(vlaue.MovieId), "GET", "", nil, nil, nil, nil)
				sp.AddRequest(req)
			}
		} else {
			req := request.NewRequest(fmt.Sprintf(baiduBaikeUrl, vlaue.MovieTitle), "html", strconv.Itoa(vlaue.MovieId), "GET", "", nil, nil, nil, nil)
			sp.AddRequest(req)
		}
	}
	sp.SetThreadnum(threadnum)
	sp.Run()
	this.Ctx.WriteString("Start Spider!")
}

var baikeDict = make(map[string]string)

func init() {
	baikeDict["中文名"] = "MovieTitleCN"
	baikeDict["外文名"] = "MovieTitleEN"
	baikeDict["其它译名"] = "MovieTitleOther"
	baikeDict["出品时间"] = "ProductionTime"
	baikeDict["出品公司"] = "ProductionCompany"
	baikeDict["发行公司"] = "DistributionFirm"
	baikeDict["制片地区"] = "ProductionArea"
	baikeDict["制片成本"] = "ProductionCost"
	baikeDict["拍摄地点"] = "ShootingLocation"
	baikeDict["拍摄日期"] = "ShootingDate"
	baikeDict["导演"] = "Director"
	baikeDict["编剧"] = "ScreenWriter"
	baikeDict["制片人"] = "Producer"
	baikeDict["类型"] = "MovieStyle"
	baikeDict["主演"] = "Starring"
	baikeDict["片长"] = "RunningTime"
	baikeDict["上映时间"] = "ReleaseTime"
	baikeDict["票房"] = "BoxOffice"
	baikeDict["对白语言"] = "DialogueLanguage"
}
