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
	"strconv"
	"strings"
)

const (
	xuandySearchUrl         = "http://www.xuandy.com/index.php?s=%s&submit=站内搜索"
	xuandyMovieUrl          = "http://www.xuandy.com/category/movie"
	xuandyMoviePageUrl      = "http://www.xuandy.com/CATEGORY/MOVIE/page/%d"
	xuandyTelevisionUrl     = "http://www.xuandy.com/category/television"
	xuandyTelevisionPageUrl = "http://www.xuandy.com/category/television/page/%d"
	xuandyVideoUrl          = "http://www.xuandy.com/category/Video"
	xuandyVideoPageUrl      = "http://www.xuandy.com/category/Video/page/%d"
	getPage                 = true
	moviePageNumber         = 600
	televisionPageNumber    = 300
	videoPageNumber         = 33
	synchPage               = 5
)

type XuandyProcesser struct {
}

func NewXuandyProcesser() *XuandyProcesser {
	return &XuandyProcesser{}
}

func getXuandyDetailUrl(p *page.Page, query *goquery.Document, urlTag string) {
	query.Find(`div[id="center"] div[class="postlist"]`).Each(func(i int, s *goquery.Selection) {
		a := s.Find(`h4 a`)
		url, isExsit := a.Attr("href")
		if isExsit {
			req := request.NewRequest(url, "html", urlTag, "GET", "", nil, nil, nil, a.Text())
			p.AddTargetRequestWithParams(req)
		}
	})
}
func getXuandyTitle(content *goquery.Selection) string {
	index := strings.Split(content.Find("h2").Text(), "》")

	title := strings.Split(index[0], "《")
	if len(title) == 2 {
		return title[1]
	} else {
		return content.Find("h2").Text()
	}
}
func (this *XuandyProcesser) initMovie(content *goquery.Selection, movie *models.Movie, p *page.Page, movieType string) {
	movie.MovieType = movieType
	reqUrl := regNumber.FindString(p.GetRequest().GetUrl())
	xuandyId, _ := strconv.ParseInt(reqUrl, 10, 64)
	movie.XuandyId = int(xuandyId)
	movie.MovieTitle = getXuandyTitle(content)
}

func (this *XuandyProcesser) getMovieDetail(query *goquery.Document, p *page.Page) {
	content := query.Find(`div[id="center"] div[class="post"]`)

	movie := models.Movie{}
	this.initMovie(content, &movie, p, "Movie")

	entry := content.Find(`div[class="entry"] p`)
	startDownloadUrl := false
	urls := make(map[string]string)
	isExsit := false
	entry.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			movie.MovieSummary = s.Text()
		} else if i == 1 {
			movie.MovieImgSrc, isExsit = s.Find("img").Attr("src")
			if !isExsit {
				movie.MovieImgSrc = ""
			}

		} else if startDownloadUrl {
			a := s.Find("a")
			href, isExsit := a.Attr("href")
			if isExsit {
				urls[s.Text()] = href
			}

		} else {
			str := s.Text()
			downloadStr := "下载地址："

			if strings.Contains(str, downloadStr) {
				startDownloadUrl = true
			} else {
				if strings.Contains(str, "◎片 名") {
					fields := strings.Split(str, "◎")
					for _, field := range fields {
						if strings.Contains(field, "译 名") {
							movie.MovieTitleOther = strings.Replace(field, "译 名", "", 1)
						} else if strings.Contains(field, "片 名") {
							movie.MovieTitleCN = strings.Replace(field, "片 名", "", 1)
						} else if strings.Contains(field, "年 代") {
							movie.ProductionTime = strings.Replace(field, "年 代", "", 1)
						} else if strings.Contains(field, "国 家") {
							movie.ProductionArea = strings.Replace(field, "国 家", "", 1)
						} else if strings.Contains(field, "类 别") {
							movie.MovieStyle = strings.Replace(field, "类 别", "", 1)
						} else if strings.Contains(field, "语 言") {
							movie.DialogueLanguage = strings.Replace(field, "语 言", "", 1)
						} else if strings.Contains(field, "片 长") {
							movie.RunningTime = strings.Replace(field, "片 长", "", 1)
						} else if strings.Contains(field, "导 演") {
							movie.Director = strings.Replace(field, "导 演", "", 1)
						} else if strings.Contains(field, "主 演 ") {
							movie.Starring = strings.Replace(field, "主 演", "", 1)
						}
					}
				}

			}
		}

	})
	this.insertMovie(&movie, urls)
}
func (this *XuandyProcesser) insertSynchError(movieTitle, errMgs string, xuandyId int, o orm.Ormer) {
	synchError := models.SynchError{}
	synchError.MovieTitle = movieTitle
	synchError.XuandyId = xuandyId
	synchError.ErrorMgs = errMgs
	o.Insert(synchError)
}
func (this *XuandyProcesser) insertXuandy(xuandyId int, url, movieType string) {
	o := orm.NewOrm()
	if o.QueryTable("Xuandy").Filter("XuandyId", xuandyId).Exist() {
		fmt.Println("Already Exists!")
		return
	}

	xuandy := models.Xuandy{}
	xuandy.XuandyId = xuandyId
	xuandy.Url = url
	xuandy.MovieType = movieType
	_, err := o.Insert(xuandy)
	fmt.Println(xuandy)
	if err != nil {
		fmt.Println("Insert error:", err)
	}
}
func (this *XuandyProcesser) insertMovie(movie *models.Movie, urls map[string]string) {
	o := orm.NewOrm()
	if movie.MovieTitle == "" {
		//insertSynchError(movie.MovieTitle, "Insert Movie error:Title is null", movie.XuandyId, o)
		return
	}
	var temp models.Movie
	var id int64
	err := o.QueryTable("Movie").Filter("XuandyId", movie.XuandyId).One(&temp)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		//insertSynchError(movie.MovieTitle, "Returned Multi Rows Not One:"+err.Error(), movie.XuandyId, o)
		return
	} else if err == orm.ErrNoRows {
		// 没有找到记录
		id, err = o.Insert(movie)
	} else {
		movie.MovieId = temp.MovieId
		id = int64(movie.MovieId)
		// _, err = o.Update(movie)
	}

	if err != nil {
		//insertSynchError(movie.MovieTitle, "Insert Movie error:"+err.Error(), movie.XuandyId, o)
		//fmt.Println("Insert Movie error:", err, ",Movie Name", movie.MovieTitle, ",XuandyId=", movie.XuandyId)
	} else {
		qs := o.QueryTable("MovieDownloadUrl").Filter("MovieId", int(id))

		for key, url := range urls {
			if strings.Trim(url, " ") == "" {
				continue
			}
			if qs.Filter("DownloadUrl", url).Exist() {
				continue
			}
			downloadUrl := models.MovieDownloadUrl{}
			downloadUrl.MovieId = int(id)
			downloadUrl.DownloadUrl = url
			downloadUrl.UrlTitle = key
			_, err := o.Insert(&downloadUrl)
			if err != nil {
				//insertSynchError(movie.MovieTitle, "Insert Download Url error:"+err.Error(), movie.XuandyId, o)
				//fmt.Println("Insert Download Url error:", err, movie.MovieTitle, ",XuandyId=", movie.XuandyId)
			}
		}
	}
}
func (this *XuandyProcesser) getTelevisionDetail(query *goquery.Document, p *page.Page) {
	content := query.Find(`div[id="center"] div[class="post"]`)
	movie := models.Movie{}
	this.initMovie(content, &movie, p, "Television")

	entry := content.Find(`div[class="entry"] p`)
	isExsit := false
	entry.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			movie.MovieSummary = s.Text()
		} else if i == 1 {
			movie.MovieImgSrc, isExsit = s.Find("img").Attr("src")
			if !isExsit {
				movie.MovieImgSrc = ""
			}
		} else {
			str := s.Text()

			if strings.Contains(str, "剧名：") {
				params := strings.Split(str, "\n")
				for _, field := range params {
					if strings.Contains(field, "剧名：") {
						movie.MovieTitleCN = strings.Replace(field, "名：", "", 1)
					} else if strings.Contains(field, "类型：") {
						movie.MovieStyle = strings.Replace(field, "类型：", "", 1)
					} else if strings.Contains(field, "地区：") {
						movie.ProductionArea = strings.Replace(field, "地区：", "", 1)
					} else if strings.Contains(field, "上映年代：") {
						movie.ProductionTime = strings.Replace(field, "上映年代：", "", 1)
					} else if strings.Contains(field, "主演：") {
						movie.Starring = strings.Replace(field, "主演：", "", 1)
					}
				}
			}
		}
	})
	urls := strings.Split(content.Find(`div[class="entry"] ol`).Text(), "\n")
	this.insertMovie2(&movie, urls)
}
func (this *XuandyProcesser) insertMovie2(movie *models.Movie, urls []string) {
	o := orm.NewOrm()
	if movie.MovieTitle == "" {
		// insertSynchError(movie.MovieTitle, "Insert Movie error:Title is null", movie.XuandyId, o)
		return
	}

	var temp models.Movie
	var id int64
	err := o.QueryTable("Movie").Filter("XuandyId", movie.XuandyId).One(&temp)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		fmt.Printf("Returned Multi Rows Not One")
		return
	} else if err == orm.ErrNoRows {
		// 没有找到记录
		id, err = o.Insert(movie)
	} else {

		movie.MovieId = temp.MovieId
		id = int64(movie.MovieId)
		// _, err = o.Update(movie)
	}

	if err != nil {
		// insertSynchError(movie.MovieTitle, "Insert Movie error:"+err.Error(), movie.XuandyId, o)
		// fmt.Println("Insert Television error:", err, ",Movie Name", movie.MovieTitle, ",XuandyId=", movie.XuandyId)
	} else {
		qs := o.QueryTable("MovieDownloadUrl").Filter("MovieId", int(id))
		for _, url := range urls {
			if strings.Trim(url, " ") == "" {
				continue
			}
			if qs.Filter("DownloadUrl", url).Exist() {
				continue
			}
			downloadUrl := models.MovieDownloadUrl{}
			downloadUrl.MovieId = int(id)
			downloadUrl.DownloadUrl = url
			downloadUrl.UrlTitle = url
			_, err := o.Insert(&downloadUrl)
			if err != nil {
				// insertSynchError(movie.MovieTitle, "Insert Download Url error:"+err.Error(), movie.XuandyId, o)
				// fmt.Println("Insert Download Url error:", err, movie.MovieTitle, ",XuandyId=", movie.XuandyId)
			}
		}
	}
}
func (this *XuandyProcesser) getVideoDetail(query *goquery.Document, p *page.Page) {
	content := query.Find(`div[id="center"] div[class="post"]`)
	movie := models.Movie{}
	this.initMovie(content, &movie, p, "Video")

	entry := content.Find(`div[class="entry"] p`)
	urls := make(map[string]string)
	isExsit, startDownloadUrl, isUrl := false, false, false
	entry.Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			movie.MovieSummary = s.Text()
		} else if i == 1 {
			movie.MovieImgSrc, isExsit = s.Find("img").Attr("src")
			if !isExsit {
				movie.MovieImgSrc = ""
			}

		} else if startDownloadUrl {
			a := s.Find("a")
			href, isExsit := a.Attr("href")
			if isExsit {
				urls[s.Text()] = href
			}

		} else if isUrl {
			urls[s.Text()] = s.Text()
		} else {
			str := s.Text()
			downloadStr := "下载地址："
			downloadStr2 := "下载方法"
			if strings.Contains(str, downloadStr) {
				startDownloadUrl = true
			} else if strings.Contains(s.Find("strong").Text(), downloadStr2) {
				isUrl = true
			} else {
				if strings.Contains(str, "剧名：") {
					params := strings.Split(str, "\n")
					for _, field := range params {
						if strings.Contains(field, "剧名：") {
							movie.MovieTitleCN = strings.Replace(field, "名：", "", 1)
						} else if strings.Contains(field, "类型：") {
							movie.MovieStyle = strings.Replace(field, "类型：", "", 1)
						} else if strings.Contains(field, "地区：") {
							movie.ProductionArea = strings.Replace(field, "地区：", "", 1)
						} else if strings.Contains(field, "上映年代：") {
							movie.ProductionTime = strings.Replace(field, "上映年代：", "", 1)
						} else if strings.Contains(field, "主演：") {
							movie.Starring = strings.Replace(field, "主演：", "", 1)
						}
					}
				}

			}
		}

	})
	this.insertMovie(&movie, urls)
}

func (this *XuandyProcesser) Process(p *page.Page) {
	query := p.GetHtmlParser()
	urlTag := p.GetUrlTag()
	if urlTag == "Movie" {
		getXuandyDetailUrl(p, query, "Movie Detail")
	} else if urlTag == "Movie Detail" {
		this.getMovieDetail(query, p)
	} else if urlTag == "Television" {
		getXuandyDetailUrl(p, query, "Television Detail")
	} else if urlTag == "Television Detail" {
		this.getTelevisionDetail(query, p)
	} else if urlTag == "Video" {
		getXuandyDetailUrl(p, query, "Video Detail")
	} else if urlTag == "Video Detail" {
		this.getVideoDetail(query, p)
	}
}

type SpiderController struct {
	base.BaseController
}

func (this *SpiderController) SpiderXuandy(moviePage, televisionPage, videoPage int) {
	movieSpider := spider.NewSpider(NewXuandyProcesser(), "xuandy")
	scheduler := myspider.NewRedisScheduler()
	movieSpider.SetScheduler(scheduler)
	movieSpider.SetSleepTime("rand", 50, 1000)
	movieSpider.SetThreadnum(threadnum)
	//电影
	movieReq := request.NewRequest(xuandyMovieUrl, "html", "Movie", "GET", "", nil, nil, nil, nil)
	movieSpider.AddRequest(movieReq)
	if getPage {
		for i := 2; i < moviePage; i++ {
			req := request.NewRequest(fmt.Sprintf(xuandyMoviePageUrl, i), "html", "Movie", "GET", "", nil, nil, nil, nil)
			movieSpider.AddRequest(req)

		}
	}
	//电视剧
	televisionReq := request.NewRequest(xuandyTelevisionUrl, "html", "Television", "GET", "", nil, nil, nil, nil)
	movieSpider.AddRequest(televisionReq)
	if getPage {
		for i := 2; i < televisionPage; i++ {
			req := request.NewRequest(fmt.Sprintf(xuandyTelevisionPageUrl, i), "html", "Television", "GET", "", nil, nil, nil, nil)
			movieSpider.AddRequest(req)
		}
	}
	//视频
	videoReq := request.NewRequest(xuandyVideoUrl, "html", "Video", "GET", "", nil, nil, nil, nil)
	movieSpider.AddRequest(videoReq)
	if getPage {
		for i := 2; i < videoPage; i++ {
			req := request.NewRequest(fmt.Sprintf(xuandyVideoPageUrl, i), "html", "Video", "GET", "", nil, nil, nil, nil)
			movieSpider.AddRequest(req)
		}
	}
	movieSpider.Run()
	this.Ctx.WriteString("Start Spider!")
}
func (this *SpiderController) SpiderXuandyAll() {
	this.SpiderXuandy(moviePageNumber, televisionPageNumber, videoPageNumber)
}
func (this *SpiderController) SynchXuandy() {
	this.SpiderXuandy(5, 5, 5)
}

type SynchDownloadUrlProcesser struct {
}

func (this *SynchDownloadUrlProcesser) Process(p *page.Page) {
	urlTag := p.GetUrlTag()
	query := p.GetHtmlParser()
	if strings.HasPrefix(urlTag, "search") {
		str := strings.Split(p.GetUrlTag(), "|@|")
		getXuandyDetailUrl(p, query, str[1])
	} else {
		o := orm.NewOrm()
		movie := models.Movie{}
		movieId, err := strconv.ParseInt(urlTag, 10, 32)

		if err != nil {
			return
		}
		movie.MovieId = int(movieId)
		o.Read(&movie)
		content := query.Find(`div[id="center"] div[class="post"]`)
		title := getXuandyTitle(content)
		if title == movie.MovieTitle || strings.Replace(title, "电影版 ", "", 1) == movie.MovieTitle {

			url := p.GetRequest().GetUrl()
			reqUrl := regNumber.FindString(url)
			xuandyId, _ := strconv.ParseInt(reqUrl, 10, 64)
			movie.XuandyId = int(xuandyId)
			url = strings.ToLower(url)
			fmt.Println(movie.MovieType)
			if strings.Contains(url, "/movie/") && strings.ToLower(movie.MovieType) == "movie" {

				entry := content.Find(`div[class="entry"] p`)
				startDownloadUrl := false
				urls := make(map[string]string)
				isExsit := false
				entry.Each(func(i int, s *goquery.Selection) {
					if i == 0 {
						movie.MovieSummary = s.Text()
					} else if i == 1 {
						movie.MovieImgSrc, isExsit = s.Find("img").Attr("src")
						if !isExsit {
							movie.MovieImgSrc = ""
						}

					} else if startDownloadUrl {
						a := s.Find("a")
						href, isExsit := a.Attr("href")
						if isExsit {
							urls[s.Text()] = href
						}

					} else {
						str := s.Text()
						downloadStr := "下载地址："

						if strings.Contains(str, downloadStr) {
							startDownloadUrl = true
						}
					}
				})

				o.Update(&movie, "MovieImgSrc", "XuandyId")

				fmt.Println(title)
				//
				qs := o.QueryTable("MovieDownloadUrl").Filter("MovieId", movie.MovieId)
				for key, url := range urls {
					if strings.Trim(url, " ") == "" {
						continue
					}
					if qs.Filter("DownloadUrl", url).Exist() {
						continue
					}
					downloadUrl := models.MovieDownloadUrl{}
					downloadUrl.MovieId = movie.MovieId
					downloadUrl.DownloadUrl = url
					downloadUrl.UrlTitle = key
					o.Insert(&downloadUrl)
				}

			} else if strings.Contains(url, "/television/") && strings.ToLower(movie.MovieType) == "tv" {

				urls := strings.Split(content.Find(`div[class="entry"] ol`).Text(), "\n")
				qs := o.QueryTable("MovieDownloadUrl").Filter("MovieId", movie.MovieId)
				o.Update(&movie, "MovieImgSrc", "XuandyId")

				for _, url := range urls {
					if strings.Trim(url, " ") == "" {
						continue
					}
					if qs.Filter("DownloadUrl", url).Exist() {
						continue
					}
					downloadUrl := models.MovieDownloadUrl{}
					downloadUrl.MovieId = movie.MovieId
					downloadUrl.DownloadUrl = url
					downloadUrl.UrlTitle = url
					o.Insert(&downloadUrl)
				}
			} else if strings.Contains(url, "/video/") && strings.ToLower(movie.MovieType) == "video" {
				entry := content.Find(`div[class="entry"] p`)
				urls := make(map[string]string)
				isExsit, startDownloadUrl, isUrl := false, false, false
				entry.Each(func(i int, s *goquery.Selection) {
					if i == 0 {
						movie.MovieSummary = s.Text()
					} else if i == 1 {
						movie.MovieImgSrc, isExsit = s.Find("img").Attr("src")
						if !isExsit {
							movie.MovieImgSrc = ""
						}

					} else if startDownloadUrl {
						a := s.Find("a")
						href, isExsit := a.Attr("href")
						if isExsit {
							urls[s.Text()] = href
						}

					} else if isUrl {
						urls[s.Text()] = s.Text()
					} else {
						str := s.Text()
						downloadStr := "下载地址："
						downloadStr2 := "下载方法"
						if strings.Contains(str, downloadStr) {
							startDownloadUrl = true
						} else if strings.Contains(s.Find("strong").Text(), downloadStr2) {
							isUrl = true
						}
					}

				})
				if movie.IsSynchBaidu || movie.IsSynchDouban {
					o.Update(movie, "MovieImgSrc", "XuandyId")
				} else {
					o.Update(movie, "MovieImgSrc", "MovieSummary", "XuandyId")
				}
				//
				qs := o.QueryTable("MovieDownloadUrl").Filter("MovieId", movie.MovieId)
				for key, url := range urls {
					if strings.Trim(url, " ") == "" {
						continue
					}
					if qs.Filter("DownloadUrl", url).Exist() {
						continue
					}
					downloadUrl := models.MovieDownloadUrl{}
					downloadUrl.MovieId = movie.MovieId
					downloadUrl.DownloadUrl = url
					downloadUrl.UrlTitle = key
					o.Insert(&downloadUrl)
				}
			}
		}
	}
}
func (this *SpiderController) SynchDownloadUrl() {
	sp := spider.NewSpider(&SynchDownloadUrlProcesser{}, "SynchDownloadUrl")
	sp.SetSleepTime("rand", 50, 200)
	sp.SetThreadnum(threadnum)
	o := orm.NewOrm()
	var movies []models.Movie
	qs := o.QueryTable("Movie").Filter("XuandyId", 0).Limit(-1)
	models.ListObjects(qs, &movies)
	for _, vlaue := range movies {
		req := request.NewRequest(fmt.Sprintf(xuandySearchUrl, vlaue.MovieTitle), "html", "search|@|"+strconv.Itoa(vlaue.MovieId), "GET", "", nil, nil, nil, nil)
		sp.AddRequest(req)
	}
	sp.Run()
	this.Ctx.WriteString("End Spider!")
}
