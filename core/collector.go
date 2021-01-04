 package core

 import (
	 "crypto/tls"
	 "fmt"
	 "github.com/Chain-Zhang/pinyin"
	 "github.com/PuerkitoBio/goquery"
	 "github.com/parnurzeal/gorequest"
	 "github.com/polaris1119/keyword"
	 "github.com/robfig/cron/v3"
	 "golang.org/x/net/html/charset"
	 "lajiCollect/app/provider"
	 "lajiCollect/config"
	 constant "lajiCollect/config/constant/db"
	 "lajiCollect/library"
	 "lajiCollect/model"
	 "lajiCollect/services"
	 "log"
	 "net/http"
	 "net/url"
	 "regexp"
	 "strconv"
	 "strings"
	 "sync"
	 "time"
	 "unicode/utf8"
 )

type RequestData struct {
	Body   string
	Domain string
	Scheme string
	IP     string
	Server string
}

var waitGroup sync.WaitGroup
var ch chan string

func Crond() {
	//一次使用几个通道
	ch = make(chan string, config.CollectorConfig.Channels)

	keyword.Extractor.Init(keyword.DefaultProps, true, config.ExecPath+"dictionary.txt")

	//1小时运行一次，采集地址，加入到地址池

	//每分钟运行一次，检查是否有需要采集的文章s
	crontab := cron.New(cron.WithSeconds())
	//10分钟抓一次列表
	crontab.AddFunc("1 */10 * * * *", CollectListTask)
	//1分钟抓一次详情
	crontab.AddFunc("1 */1 * * * *", CollectDetailTask)
	crontab.Start()
	//启动的时候，先执行一遍
	//go CollectListTask()
	go CollectDetailTask()
}

//采集列表
func CollectListTask() {
	if services.DB == nil {
		return
	}
	fmt.Println("collect list")
	db := services.DB
	var articleSources []model.ArticleSource
	err := db.Model(model.ArticleSource{}).Where("`error_times` < ? AND is_monitor=?", config.CollectorConfig.ErrorTimes,1).Find(&articleSources).Error
	if err != nil {
		return
	}

	for _, v := range articleSources {
		getArticleLinks(v)
	}
}

//采集文章详情
func CollectDetailTask() {
	if services.DB == nil {
		return
	}
	fmt.Println("collect detail")
	//检查article的地址
	var articleList []model.Article

	db := services.DB
	db.Debug().Model(model.Article{}).Where("status = 0").Order("id asc").Limit(config.CollectorConfig.Channels * 100).Scan(&articleList)
	for _, vv := range articleList {
		ch <- vv.OriginUrl
		waitGroup.Add(1)
		go getArticleDetail(vv)
	}

	waitGroup.Wait()
}

//根据列表获取链接
func getArticleLinks(v model.ArticleSource) {
	GetArticleLinks(&v)
}

//获取文章详情 加入队列
func getArticleDetail(v model.Article) {
	 defer func() {
		 waitGroup.Done()
		 <-ch
	 }()

	 GetArticleDetail(&v)
 }

 //wordpress网站Rss链接
 func GetArticleDetailWordpressRss(v *model.ArticleSource) error {
	 requestData, err := Request(v.Url)
	 if err != nil {
		 log.Println(err)
		 return err
	 }

	 requestData.Body = strings.ReplaceAll(requestData.Body,"content:encoded","contentEncoded")
	 requestData.Body = strings.ReplaceAll(requestData.Body,"dc:creator","dcCreator")
	 requestData.Body = strings.ReplaceAll(requestData.Body,"<![CDATA[","")
	 requestData.Body = strings.ReplaceAll(requestData.Body,"]]>","")
	 requestData.Body = strings.ReplaceAll(requestData.Body,"<link>","<linkR>")
	 requestData.Body = strings.ReplaceAll(requestData.Body,"</link>","</linkR>")
	 htmlR := strings.NewReader(requestData.Body)
	 doc, err := goquery.NewDocumentFromReader(htmlR)
	 if err != nil {
		 return err
	 }
	 items := doc.Find("item")
	 for i := range items.Nodes{
	 	 nowEq := items.Eq(i)
	 	 article := &model.Article{}
		 article.SourceId 		= v.Id
		 article.ArticleType 	= v.UrlType
		 article.Status 		= constant.DbArticleStatusPass
		 article.Title 			= nowEq.Find("title").Text()
		 article.OriginUrl    = nowEq.Find("linkR").Text()
		 article.Author 		= nowEq.Find("dcCreator").Text()
		 article.Description 	= nowEq.Find("description").Text()
		 //TODO 把分类的也加入关键词
		 article.Keywords = nowEq.Find("category").Text()
		 article.Keywords = strings.ReplaceAll(article.Keywords," ",",")
		 keywords 		 := library.GetKeywords(article.Title, 5)
		 article.Keywords = article.Keywords+strings.Join(keywords, ",")
		 //内容
		 html,_		:= nowEq.Find("contentEncoded").Html()
		 article.Content = article.FormatContent(html,v)
		 article.Save()
	 }
	 return nil
 }


//采集链接
func GetArticleLinks(v *model.ArticleSource) {

	db := services.DB
	switch v.UrlType {
		case constant.DbArticleUrlTypeWordpressRss:
				//TODO wordpress Rss特殊情况
				err := GetArticleDetailWordpressRss(v)
				if err != nil {
					goto RETURNERR
				}
				return
			break
	case constant.DbArticleUrlTypeDetail:
		//先检查数据库里有没有，没有的话，就抓回来
		article := &model.Article{}
		article.CreatedTime 	= int(time.Now().Unix())
		article.SourceId 		= v.Id
		article.ArticleType 	= v.UrlType
		article.Status 			= 0
		article.OriginUrl 		= v.Url
		db.Model(model.Article{}).Where(model.Article{OriginUrl: article.OriginUrl}).FirstOrCreate(&article)
		return
		break
	case constant.DbArticleUrlTypeList:
	default:
		urlParse, err := url.Parse(v.Url)
		if err == nil {
			articleList, err := CollectLinks(v.Url)
			if err != nil {
				goto RETURNERR
			}

			for _, article := range articleList {

				rule,err := v.GetParseRule()
				if err == nil {
					//判断是否只拿属于该网站的链接
					if rule.UrlOnlySelf == 1 {
						articleUrlParse, err := url.Parse(article.OriginUrl)
						if err != nil {
							continue
						}

						if urlParse.Host != articleUrlParse.Host {
							continue
						}
					}
				}

				//先检查数据库里有没有，没有的话，就抓回来
				article.CreatedTime = int(time.Now().Unix())
				article.SourceId = v.Id
				article.ArticleType = v.UrlType
				article.Status = 0
				db.Model(model.Article{}).Where(model.Article{OriginUrl: article.OriginUrl}).FirstOrCreate(&article)
			}
			return
		} else {
			goto RETURNERR
		}
	}
	RETURNERR:
		db.Model(v).Update("error_times", v.ErrorTimes+1)
}

//万能抓取详情页
func GetArticleDetail(v *model.Article) {
	db := services.DB
	//标记当前为执行中
	db.Model(model.Article{}).Where("`id` = ?", v.Id).Update("status", 2)

	//开始抓取详情
	_ = CollectDetail(v)

	//更新到数据库中
	status := int(1)
	if v.Content == "" {
		status = 3
	}
	if utf8.RuneCountInString(v.Title) < 10 {
		status = 3
	}
	urlArr := strings.Split(v.OriginUrl, "/")
	if len(urlArr) <= 3 {
		status = 3
	}
	if len(urlArr) <= 4 && strings.HasPrefix(v.OriginUrl, "/") {
		status = 3
	}

	if strings.Contains(v.Title, "法律声明") || strings.Contains(v.Title, "关于我们") || strings.Contains(v.Title, "站点地图") || strings.Contains(v.Title, "区长信箱") || strings.Contains(v.Title, "政务服务网") || strings.Contains(v.Title, "政务公开") || strings.Contains(v.Title, "人民政府网站") || strings.Contains(v.Title, "门户网站") || strings.Contains(v.Title, "领导介绍") || strings.Contains(v.Title, "403") || strings.Contains(v.Title, "404") || strings.Contains(v.Title, "Government") || strings.Contains(v.Title, "China") {
		status = 3
	}
	//小于500字 内容，不过审
	if utf8.RuneCountInString(v.ContentText) < 200 {
		status = 3
	}
	if strings.Contains(v.ContentText, "ICP备") || strings.Contains(v.ContentText, "政府网站标识码") || strings.Contains(v.ContentText, "以上版本浏览本站") || strings.Contains(v.ContentText, "版权声明") || strings.Contains(v.ContentText, "公网安备") {
		status = 3
	}

	db.Model(model.Article{}).Where("`id` = ?", v.Id).Update("status", status)

	timeTemplate1 := "2006-01-02 15:04:05"
	timestamp := int(time.Now().Unix())
	pubTime, _ := time.ParseInLocation(timeTemplate1, v.PubDate, time.Local)
	if pubTime.Unix() > 0 {
		timestamp = int(pubTime.Unix())
	}

	v.UpdatedTime = int(time.Now().Unix())
	v.CreatedTime = timestamp
	v.Status = status

	article := v
	fmt.Println(status, v.Title, v.OriginUrl)
	article.Save()

	AutoPublish(article)
}

//自动发布推送
func AutoPublish(article *model.Article) {
	if config.ContentConfig.AutoPublish == 0 || article.Status != 1 {
		return
	}
	publishData := map[string]string{
		config.ContentConfig.TitleField: article.Title,
	}
	if config.ContentConfig.KeywordsField != "" {
		publishData[config.ContentConfig.KeywordsField] = article.Keywords
	}
	if config.ContentConfig.DescriptionField != "" {
		publishData[config.ContentConfig.DescriptionField] = article.Description
	}
	if config.ContentConfig.CreatedTimeField != "" {
		publishData[config.ContentConfig.CreatedTimeField] = strconv.Itoa(article.CreatedTime)
	}
	if config.ContentConfig.AuthorField != "" {
		publishData[config.ContentConfig.AuthorField] = article.Author
	}
	if config.ContentConfig.ViewsField != "" {
		publishData[config.ContentConfig.ViewsField] = strconv.Itoa(article.Views)
	}
	if config.ContentConfig.TableName == config.ContentConfig.ContentTableName || config.ContentConfig.ContentTableName == "" || config.ContentConfig.AutoPublish == 2 {
		if config.ContentConfig.ContentField != "" {
			publishData[config.ContentConfig.ContentField] = article.Content
		}
	}
	if len(config.ContentConfig.ExtraFields) > 0 {
		for _, v := range config.ContentConfig.ExtraFields {
			value := v.Value
			if v.Value == "{id}" {
				//获取id
				value = strconv.Itoa(article.Id)
			} else if v.Value == "{py}" {
				//获取标题首字母
				str, err := pinyin.New(article.Title).Split("-").Mode(pinyin.WithoutTone).Convert()
				if err == nil {
					value = ""
					strArr := strings.Split(str, "-")
					for _, v := range strArr {
						value += string(v[0])
					}
				}
			} else if v.Value == "{pinyin}" {
				//获取标题拼音
				str, err := pinyin.New(article.Title).Split("").Mode(pinyin.WithoutTone).Convert()
				if err == nil {
					value = str
				}
			} else if v.Value == "{time}" {
				//获取标题首字母
				value = strconv.Itoa(int(time.Now().Unix()))
			} else if v.Value == "{date}" {
				//获取标题首字母
				value = time.Now().Format("2006-01-02")
			}
			publishData[v.Key] = value
		}
	}

	if config.ContentConfig.AutoPublish == 1 {
		//本地发布
		publishDataKeys := make([]string, len(publishData))
		publishDataValues := make([]string, len(publishData))
		j := 0
		for k, v := range publishData {
			publishDataKeys[j] = k
			publishDataValues[j] = fmt.Sprintf("'%s'", v)
			j++
		}

		insertId := int64(0)
		result, err := services.DB.DB().Exec(fmt.Sprintf("INSERT INTO `%s` (%s)VALUES(%s)", config.ContentConfig.TableName, strings.Join(publishDataKeys, ","), strings.Join(publishDataValues, ",")))
		if err == nil {
			insertId, err = result.LastInsertId()
			if config.ContentConfig.ContentTableName != "" && config.ContentConfig.TableName != config.ContentConfig.ContentTableName {
				services.DB.Exec(fmt.Sprintf("INSERT INTO `%s` (%s, %s)VALUES(?, ?)", config.ContentConfig.ContentTableName, config.ContentConfig.ContentIdField, config.ContentConfig.ContentField), insertId, article.Content)
			}
		}
	} else if config.ContentConfig.AutoPublish == 2 && config.ContentConfig.RemoteUrl != "" {
		//headers
		sg := gorequest.New().Timeout(10 * time.Second).Post(config.ContentConfig.RemoteUrl)

		//判断请求内容类型
		if config.ContentConfig.ContentType == "json" {
			sg = sg.Set("Content-Type", "multipart/form-data")
		} else if config.ContentConfig.ContentType == "urlencode" {
			sg = sg.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			sg = sg.Set("Content-Type", "application/json")
		}

		//加上请求头
		if len(config.ContentConfig.Headers) > 0 {
			for _, v := range config.ContentConfig.Headers {
				sg = sg.Set(v.Key, v.Value)
			}
		}

		//加上cookie
		if len(config.ContentConfig.Cookies) > 0 {
			urlInfo, _ := url.Parse(config.ContentConfig.RemoteUrl)
			for _, v := range config.ContentConfig.Cookies {
				cookie := &http.Cookie{
					Name:    v.Key,
					Value:   v.Value,
					Path:    "/",
					Domain:  urlInfo.Hostname(),
					Expires: time.Now().Add(86400 * time.Second),
				}
				sg = sg.AddCookie(cookie)
			}
		}

		//不接收处理结果
		resp, _, errs := sg.SendMap(publishData).End()
		if len(errs) > 0 {
			fmt.Println(errs)
			return
		}
		defer resp.Body.Close()
		fmt.Println("请求发布结果")
		library.DEBUG(resp.Body)
	}
}

//抓取链接
func CollectLinks(link string) ([]model.Article, error) {
	requestData, err := Request(link)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	htmlR := strings.NewReader(requestData.Body)
	doc, err := goquery.NewDocumentFromReader(htmlR)
	if err != nil {
		return nil, err
	}

	var articles []model.Article
	aLinks := doc.Find("a")
	//读取所有连接
	for i := range aLinks.Nodes {
		href, exists := aLinks.Eq(i).Attr("href")
		title := strings.TrimSpace(aLinks.Eq(i).Text())
		if exists {
			href = library.ParseLink(href, link)
		}
		if len(href) > 250 {
			href = string(href[:250])
		}
		//斜杠/结尾的抛弃
		//if strings.HasSuffix(href, "/") == false {
		articles = append(articles, model.Article{
			Title:     title,
			OriginUrl: href,
		})
		//}
	}

	return articles, nil
}

//采集详情
func CollectDetail(article *model.Article) error {
	requestData, err := Request(article.OriginUrl)
	if err != nil {
		log.Println(err)
		return err
	}
	//先删除一些不必要的标签
	re, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	requestData.Body = re.ReplaceAllString(requestData.Body, "")
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	requestData.Body = re.ReplaceAllString(requestData.Body, "")

	htmlR := strings.NewReader(requestData.Body)
	doc, err := goquery.NewDocumentFromReader(htmlR)
	if err != nil {
		return err
	}

	//获取前缀
	article.GetDomain()

	sourceInfo,_ := provider.GetArticleSourceById(article.SourceId)

	//如果是百度百科地址，单独处理
	if strings.Contains(article.OriginUrl, "baike.baidu.com") {
		article.ParseBaikeDetail(doc, requestData.Body)
	} else {
		//开始进行普通模式分析内容
		article.ParseNormalDetail(doc, requestData.Body, sourceInfo)
	}
	nameRune := []rune(article.Description)
	curLen := len(nameRune)
	if curLen > 150 {
		article.Description = string(nameRune[:150])
	}

	return nil
}

/**
 * 请求域名返回数据
 */
func Request(urlPath string) (*RequestData, error) {
	resp, body, errs := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).Timeout(90*time.Second).AppendHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36").Get(urlPath).End()
	if len(errs) > 0 {
		//如果是https,则尝试退回http请求
		if strings.HasPrefix(urlPath, "https") {
			urlPath = strings.Replace(urlPath, "https://", "http://", 1)
			return Request(urlPath)
		}
		return nil, errs[0]
	}
	defer resp.Body.Close()
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	var htmlEncode string

	if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
		htmlEncode = "gb18030"
	} else if strings.Contains(contentType, "big5") {
		htmlEncode = "big5"
	} else if strings.Contains(contentType, "utf-8") {
		htmlEncode = "utf-8"
	}
	log.Println(contentType)
	if htmlEncode == "" {
		//先尝试读取charset
		reg := regexp.MustCompile(`(?is)<meta[^>]*charset\s*=["']?\s*([A-Za-z0-9\-]+)`)
		match := reg.FindStringSubmatch(body)
		if len(match) > 1 {
			contentType = strings.ToLower(match[1])
			log.Println(contentType)
			if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
				htmlEncode = "gb18030"
			} else if strings.Contains(contentType, "big5") {
				htmlEncode = "big5"
			} else if strings.Contains(contentType, "utf-8") {
				htmlEncode = "utf-8"
			}
		}
		if htmlEncode == "" {
			reg = regexp.MustCompile(`(?is)<title[^>]*>(.*?)<\/title>`)
			match = reg.FindStringSubmatch(body)
			if len(match) > 1 {
				aa := match[1]
				_, contentType, _ = charset.DetermineEncoding([]byte(aa), "")
				log.Println(contentType)
				htmlEncode = strings.ToLower(htmlEncode)
				if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
					htmlEncode = "gb18030"
				} else if strings.Contains(contentType, "big5") {
					htmlEncode = "big5"
				} else if strings.Contains(contentType, "utf-8") {
					htmlEncode = "utf-8"
				}
			}
		}
	}
	if htmlEncode != "" && htmlEncode != "utf-8" {
		body = library.ConvertToString(body, htmlEncode, "utf-8")
	}
	log.Println(htmlEncode)

	requestData := RequestData{
		Body:   body,
		Domain: resp.Request.Host,
		Scheme: resp.Request.URL.Scheme,
		Server: resp.Header.Get("Server"),
	}

	return &requestData, nil
}

