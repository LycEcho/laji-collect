package model

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"lajiCollect/app/request"
	"lajiCollect/config"
	"lajiCollect/library"
	"lajiCollect/library/constant"
	"lajiCollect/services"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Article struct {
	Id           		int    `json:"id" gorm:"column:id;type:int(10) unsigned not null AUTO_INCREMENT;primary_key"`
	SourceId     		int    `json:"source_id" gorm:"column:source_id;type:int(11) not null COMMENT '采集源ID';default:0"`
	Title        		string `json:"title" gorm:"column:title;type:varchar(190) not null COMMENT '关键词';default:'';index:idx_title"`
	Keywords     		string `json:"keywords" gorm:"column:keywords;type:varchar(250) not null COMMENT '关键词';default:''"`
	Description  		string `json:"description" gorm:"column:description;type:varchar(250) not null COMMENT '描述';default:''"`
	Content      		string `json:"content" gorm:"-"`																								//内容
	ArticleType  		int    `json:"article_type" gorm:"column:article_type;type:tinyint(1) unsigned not null COMMENT '文章类型';default:0;index:idx_article_type"`
	OriginUrl    		string `json:"origin_url" gorm:"column:origin_url;type:varchar(250) not null COMMENT '源地址';default:'';index:idx_origin_url"`
	Author       		string `json:"author" gorm:"column:author;type:varchar(100) not null COMMENT '作者';default:''"`
	Views        		int    `json:"views" gorm:"column:views;type:int(10) not null COMMENT '查看次数';default:0;index:idx_views"`
	Status       		int    `json:"status" gorm:"column:status;type:tinyint(1) unsigned not null COMMENT '状态 【0=未发布】';default:0;index:idx_status"`
	StatusRelease       int    `json:"status_release" gorm:"column:status_release;type:tinyint(1) unsigned not null COMMENT '发布状态 【0=未发布】';default:0;index:idx_status_release"`
	CreatedTime  		int    `json:"created_time" gorm:"column:created_time;type:int(11) unsigned not null COMMENT '创建时间';default:0;index:idx_created_time"`
	UpdatedTime  		int    `json:"updated_time" gorm:"column:updated_time;type:int(11) unsigned not null COMMENT '更新时间';default:0;index:idx_updated_time"`
	DeletedTime  		int    `json:"-" gorm:"column:deleted_time;type:int(11) unsigned not null;default:0"`
	OriginDomain 		string `json:"-" gorm:"-"`
	OriginPath   		string `json:"-" gorm:"-"`
	ContentText  		string `json:"-" gorm:"-"`
	PubDate      		string `json:"-" gorm:"-"`
}

type ArticleData struct {
	Id      int    `json:"id" gorm:"column:id;type:int(10) ;unsigned not null AUTO_INCREMENT;primary_key"`
	Content string `json:"content" gorm:"column:content;type:longtext;not null;default:''"`
}

type ArticleSource struct {
	Id         int               `json:"id" gorm:"column:id;type:int(10) unsigned not null AUTO_INCREMENT;primary_key"`
	Url        string            `json:"url" gorm:"column:url;type:varchar(190) not null;default:'';index:idx_url"`
	UrlType    int               `json:"url_type" gorm:"column:url_type;type:tinyint(1) not null;default:0"`
	ErrorTimes int               `json:"error_times" gorm:"column:error_times;type:int(10) not null;default:0;index:idx_error_times"`
	Attr       *ArticleSourceAttr `json:"attr" gorm:"-"`
}

type ArticleSourceAttr struct {
	SourceId    	int    `json:"source_id" gorm:"column:source_id;type:int(10) unsigned not null;primary_key"`
	Rule 			string `json:"rule" gorm:"column:rule;type:longtext not null;"`
}

func(article *ArticleSource) GetParseRule() (*request.ArticleSourceAttrRule,error) {
	resp := &request.ArticleSourceAttrRule{}
	if article.Attr != nil {
		if err := json.Unmarshal([]byte(article.Attr.Rule), resp); err != nil{
			resp = &request.ArticleSourceAttrRule{}
			return resp,err
		}
	}
	return resp,nil
}

func (article *Article) Save(db *gorm.DB) error {
	if article.Id == 0 {
		article.CreatedTime = int(time.Now().Unix())
	}

	if err := db.Save(article).Error; err != nil {
		return err
	}
	articleData := ArticleData{
		Id:      article.Id,
		Content: article.Content,
	}
	db.Save(&articleData)

	return nil
}

func (article *Article) Delete() error {
	db := services.DB
	if err := db.Delete(article).Error; err != nil {
		return err
	}

	db.Where("id = ?", article.Id).Delete(ArticleData{})

	return nil
}

func (source *ArticleSource) Save() error {
	db := services.DB
	if err := db.Save(&source).Error; err != nil {
		return err
	}
	source.Attr.SourceId =  source.Id
	if err := db.Save(&source.Attr).Error; err != nil {
		return err
	}
	return nil
}

func (source *ArticleSource) Delete() error {
	db := services.DB
	if err := db.Delete(source).Error; err != nil {
		return err
	}

	return nil
}


//解析 百度百科
func (article *Article) ParseBaikeDetail(doc *goquery.Document, body string) {
	//获取标题
	article.Title = doc.Find("h1").Text()
	//获取描述
	reg := regexp.MustCompile(`<meta\s+name="description"\s+content="([^"]+)">`)
	match := reg.FindStringSubmatch(body)
	if len(match) > 1 {
		article.Description = match[1]
	}
	//获取关键词
	reg = regexp.MustCompile(`<meta\s+name="keywords"\s+content="([^"]+)">`)
	match = reg.FindStringSubmatch(body)
	if len(match) > 1 {
		article.Keywords = match[1]
	} else if article.Title != "" {
		keywords := library.GetKeywords(article.Title, 5)
		article.Keywords = strings.Join(keywords, ",")
	}

	doc.Find(".edit-icon").Remove()
	contentList := doc.Find(".para-title,.para")
	content := ""
	for i := range contentList.Nodes {
		content += "<p>" + contentList.Eq(i).Text() + "</p>"
	}

	article.Content = content
}

//解析正常网站
func (article *Article) ParseNormalDetail(doc *goquery.Document, body string,source *ArticleSource) {
	article.ParseTitle(doc, body)

	if article.Title != "" {
		keywords := library.GetKeywords(article.Title, 5)
		article.Keywords = strings.Join(keywords, ",")
	}

	//尝试获取正文内容
	article.ParseContent(doc, body,source)

	//尝试获取作者
	reg := regexp.MustCompile(`<meta\s+name="Author"\s+content="(.*?)"[^>]*>`)
	match := reg.FindStringSubmatch(body)
	if len(match) > 1 {
		author := match[1]
		if author == "" {
			reg := regexp.MustCompile(`(?i)(来源|作者)\s*(:|：|\s)\s*([^\s]+)`)
			match := reg.FindStringSubmatch(body)
			if len(match) > 1 {
				author = match[3]
			}
		}
		article.Author = author
	}

	//尝试获取发布时间
	reg = regexp.MustCompile(`(?i)<meta\s+name="PubDate"\s+content="(.*?)"[^>]*>`)
	match = reg.FindStringSubmatch(body)
	if len(match) > 1 {
		pubDate := match[1]
		if pubDate == "" {
			reg = regexp.MustCompile(`(?i)([0-9]{4})\s*[\-|\/|年]\s*([0-9]{1,2})\s*[\-|\/|月]\s*([0-9]{1,2})\s*([\-|\/|日])?\s*(([0-9]{1,2})\s*[:|：|时]\s*([0-9]{1,2})\s*([:|：|分])?\s*([0-9]{1,2})?)?`)
			match = reg.FindStringSubmatch(body)
			if len(match) > 1 {
				if match[1] != "" {
					pubDate = match[1] + "-" + match[2] + "-" + match[3]
				}
				if match[5] != "" {
					pubDate += " " + match[6] + ":" + match[7]
					if match[9] != "" {
						pubDate += ":" + match[9]
					} else {
						pubDate += ":00"
					}
				} else {
					pubDate += " 12:00:00"
				}
			}
		}
		article.PubDate = pubDate
	}
}

func (article *Article) ParseTitle(doc *goquery.Document, body string) {
	//尝试获取标题
	//先尝试获取h1标签
	title := ""
	h1s := doc.Find("h1")
	if h1s.Length() > 0 {
		for i := range h1s.Nodes {
			item := h1s.Eq(i)
			item.Children().Remove()
			text := strings.TrimSpace(item.Text())
			textLen := utf8.RuneCountInString(text)
			if textLen >= config.CollectorConfig.TitleMinLength && textLen > utf8.RuneCountInString(title) && !library.HasContain(text, config.CollectorConfig.TitleExclude) && !library.HasPrefix(text, config.CollectorConfig.TitleExcludePrefix) && !library.HasSuffix(text, config.CollectorConfig.TitleExcludeSuffix) {
				title = text
			}
		}
	}
	if title == "" {
		//获取 政府网站的 <meta name='ArticleTitle' content='西城法院出台案件在线办理操作指南'>
		text, exist := doc.Find("meta[name=ArticleTitle]").Attr("content")
		if exist {
			text = strings.TrimSpace(text)
			if utf8.RuneCountInString(text) >= config.CollectorConfig.TitleMinLength && !library.HasContain(text, config.CollectorConfig.TitleExclude) && !library.HasPrefix(text, config.CollectorConfig.TitleExcludePrefix) && !library.HasSuffix(text, config.CollectorConfig.TitleExcludeSuffix) {
				title = text
			}
		}
	}
	if title == "" {
		//获取title标签
		text := doc.Find("title").Text()
		text = strings.ReplaceAll(text, "_", "-")
		sepIndex := strings.Index(text, "-")
		if sepIndex > 0 {
			text = text[:sepIndex]
		}
		text = strings.TrimSpace(text)
		if utf8.RuneCountInString(text) >= config.CollectorConfig.TitleMinLength && !library.HasContain(text, config.CollectorConfig.TitleExclude) && !library.HasPrefix(text, config.CollectorConfig.TitleExcludePrefix) && !library.HasSuffix(text, config.CollectorConfig.TitleExcludeSuffix) {
			title = text
		}
	}

	log.Println(len(title), title)
	if title == "" {
		//获取title标签
		//title = doc.Find("#title,.title,.bt,.articleTit").First().Text()
		h2s := doc.Find("#title,.title,.bt,.articleTit,.right-xl>p,.biaoti")
		if h2s.Length() > 0 {
			for i := range h2s.Nodes {
				item := h2s.Eq(i)
				item.Children().Remove()
				text := strings.TrimSpace(item.Text())
				textLen := utf8.RuneCountInString(item.Text())
				if textLen >= config.CollectorConfig.TitleMinLength && textLen > utf8.RuneCountInString(title) && !library.HasContain(text, config.CollectorConfig.TitleExclude) && !library.HasPrefix(text, config.CollectorConfig.TitleExcludePrefix) && !library.HasSuffix(text, config.CollectorConfig.TitleExcludeSuffix) {
					title = text
				}
			}
		}
	}
	if title == "" {
		//如果标题为空，那么尝试h2
		h2s := doc.Find("h2,.name")
		if h2s.Length() > 0 {
			for i := range h2s.Nodes {
				item := h2s.Eq(i)
				item.Children().Remove()
				text := strings.TrimSpace(item.Text())
				textLen := utf8.RuneCountInString(text)
				if textLen >= config.CollectorConfig.TitleMinLength && textLen > utf8.RuneCountInString(title) && !library.HasContain(text, config.CollectorConfig.TitleExclude) && !library.HasPrefix(text, config.CollectorConfig.TitleExcludePrefix) && !library.HasSuffix(text, config.CollectorConfig.TitleExcludeSuffix) {
					title = text
				}
			}
		}
	}

	title = strings.Replace(strings.Replace(strings.TrimSpace(title), "\t", "", -1), "\n", " ", -1)
	title = strings.Replace(title, "<br>", "", -1)
	title = strings.Replace(title, "<br/>", "", -1)
	//只要第一个
	if utf8.RuneCountInString(title) > 50 {
		//减少误伤
		title = strings.ReplaceAll(title, "、", "-")
	}
	title = strings.ReplaceAll(title, "_", "-")
	sepIndex := strings.Index(title, "-")
	if sepIndex > 0 {
		title = title[:sepIndex]
	}

	article.Title = title
}

func (article *Article) ParseContent(doc *goquery.Document, body string,source *ArticleSource) {
	content := ""
	contentText := ""
	description := ""
	contentLength := 0

	//对一些固定的内容，直接获取值
	contentItems := doc.Find("UCAPCONTENT,#mainText,.article-content,#article-content,#articleContnet,.entry-content,.the_body,.rich_media_content,#js_content,.word_content,.pages_content,.wendang_content,#content,.RichText,.markdown-section")
	if contentItems.Length() > 0 {
		for i := range contentItems.Nodes {
			contentItem := contentItems.Eq(i)
			content, _ = contentItem.Html()
			contentText = contentItem.Text()
			contentText = strings.Replace(contentText, " ", "", -1)
			contentText = strings.Replace(contentText, "\n", "", -1)
			contentText = strings.Replace(contentText, "\r", "", -1)
			contentText = strings.Replace(contentText, "\t", "", -1)
			nameRune := []rune(contentText)
			curLen := len(nameRune)
			if curLen > 150 {
				description = string(nameRune[:150])
			}
			//判断内容的真实性
			if curLen < config.CollectorConfig.ContentMinLength {
				contentText = ""
			}
			aCount := 0
			aLinks := contentItem.Find("a")
			if aLinks.Length() > 0 {
				for i := range aLinks.Nodes {
					href, exist := aLinks.Eq(i).Attr("href")
					if exist && href != "" && !strings.HasPrefix(href, "#") {
						aCount++
					}
				}
			}
			if aCount > 5 {
				//太多连接了，直接放弃该内容
				contentText = ""
			}
			//查找内部div，如果存在，则使用它替代上一级
			divs := contentItem.Find("div")
			//只有内部没有div了或者内部div内容太少，才认为是真正的内容
			if divs.Length() > 0 {
				for i := range divs.Nodes {
					div := divs.Eq(i)
					if (div.Find("div").Length() == 0 || utf8.RuneCountInString(div.Find("div").Text()) < 100) && div.ChildrenFiltered("p").Length() > 0 && utf8.RuneCountInString(div.Text()) >= config.CollectorConfig.ContentMinLength {
						contentItem = div
						break
					}
				}
			}
			//排除一些不对的标签
			otherItems := contentItem.Find("input,textarea,form,button,footer,.footer")
			if otherItems.Length() > 0 {
				otherItems.Remove()
			}
			contentItem.Find("h1").Remove()
			//根据规则过滤
			if library.HasContain(contentText, config.CollectorConfig.ContentExclude) {
				contentText = ""
			}
			inner := contentItem.Find("*")
			for i := range inner.Nodes {
				item := inner.Eq(i)
				if library.HasContain(item.Text(), config.CollectorConfig.ContentExcludeLine) {
					item.Remove()
				}
			}

			if len(contentText) > 0 {
				break
			}
		}
	}

	if contentText == "" {
		content = ""
		//通用的获取方法
		divs := doc.Find("div,article")
		for i := range divs.Nodes {
			item := divs.Eq(i)
			pCount := item.ChildrenFiltered("p").Length()
			brCount := item.ChildrenFiltered("br").Length()
			aCount := item.Find("a").Length()
			if aCount > 5 {
				//太多连接了，直接放弃该内容
				continue
			}
			//排除一些不对的标签
			otherLength := item.Find("input,textarea,form,button,footer,.footer").Length()
			if otherLength > 0 {
				continue
			}
			if item.Find("div").Length() > 0 && utf8.RuneCountInString(item.Find("div").Text()) >= config.CollectorConfig.ContentMinLength {
				continue
			}
			if pCount > 0 || brCount > 0 {
				//表示查找到了一个p
				//移除空格和换行
				checkText := item.Text()
				checkText = strings.Replace(checkText, " ", "", -1)
				checkText = strings.Replace(checkText, "\n", "", -1)
				checkText = strings.Replace(checkText, "\r", "", -1)
				checkText = strings.Replace(checkText, "\t", "", -1)
				nameRune := []rune(checkText)
				curLen := len(nameRune)

				//根据规则过滤
				if library.HasContain(checkText, config.CollectorConfig.ContentExclude) {
					continue
				}
				if curLen <= config.CollectorConfig.ContentMinLength {
					continue
				}

				item.Find("h1,a").Remove()
				inner := item.Find("*")
				for i := range inner.Nodes {
					innerItem := inner.Eq(i)
					if library.HasContain(innerItem.Text(), config.CollectorConfig.ContentExcludeLine) {
						innerItem.Remove()
					}
				}

				if curLen > contentLength {
					contentLength = curLen
					content, _ = item.Html()
					contentText = checkText
					if curLen <= 150 {
						description = string(nameRune)
					} else {
						description = string(nameRune[:150])
					}
				}
			}
		}
	}

	//对内容进行处理
	article.ContentText = contentText
	article.Description = strings.TrimSpace(description)
	article.Content 	= article.formatContent(strings.TrimSpace(content),source)
}

//处理内容
func (article *Article) formatContent(content string,source *ArticleSource) string{

	//替换资源地址
	re, _ := regexp.Compile("src=[\"']+?(.*?)[\"']+?[^>]+?>")
	content = re.ReplaceAllStringFunc(content, article.ReplaceSrc)

	//替换链接
	re2, _ := regexp.Compile("href=[\"']+?(.*?)[\"']+?[^>]+?>")
	content = re2.ReplaceAllStringFunc(content, article.ReplaceHref)

	//清空所有className
	re3, _ := regexp.Compile(constant.RegularExpressionContentAttrClassComplete)
	content = re3.ReplaceAllLiteralString(content, "")

	//清空所有的 data-**=‘**’
	re4, _ := regexp.Compile(constant.RegularExpressionContentAttrData_Complete)
	content = re4.ReplaceAllLiteralString(content, "")


	//根据指定规则去除
	rule,err := source.GetParseRule()
	if  err == nil {
		htmlR := strings.NewReader(content)
		doc, err := goquery.NewDocumentFromReader(htmlR)
		if err == nil {
			if rule.OnlyText == 1{
				doc.Find("img,video").Remove()
			}
		}

		content, _ = doc.Html()
	}


	return content
}

func (article *Article) GetDomain() {
	baseUrlArr := strings.Split(article.OriginUrl, "/")
	pathUrlArr := baseUrlArr[:len(baseUrlArr)-1]
	baseUrlArr = baseUrlArr[:3]
	baseUrl := strings.Join(baseUrlArr, "/")
	article.OriginDomain = baseUrl
	article.OriginPath = strings.Join(pathUrlArr, "/")
}

func (article *Article) ReplaceSrc(src string) string {
	re, _ := regexp.Compile("src=[\"']+?(.*?)[\"']+?[^>]+?>")
	match := re.FindStringSubmatch(src)
	if len(match) < 1 {
		return src
	}

	if match[1] != "" {
		newSrc := library.ParseLink(match[1], article.OriginPath)
		src = strings.Replace(src, match[1], newSrc, -1)
	}
	return src
}

func (article *Article) ReplaceHref(src string) string {
	re, _ := regexp.Compile("href=[\"']+?(.*?)[\"']+?[^>]+?>")
	match := re.FindStringSubmatch(src)
	if len(match) < 1 {
		return src
	}

	if match[1] != "" {
		newSrc := library.ParseLink(match[1], article.OriginPath)
		src = strings.Replace(src, match[1], newSrc, -1)
	}
	return src
}
