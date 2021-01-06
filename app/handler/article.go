package handler

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"lajiCollect/app/provider"
	"lajiCollect/app/request"
	"lajiCollect/app/response"
	"lajiCollect/config"
	"lajiCollect/core"
	"lajiCollect/model"
	"strings"
)

func Keywords(ctx iris.Context) {
	ctx.View("article/keywords.html")
}

func ArticleSource(ctx iris.Context) {
	ctx.View("article/source.html")
}

func ArticleList(ctx iris.Context) {
	ctx.View("article/list.html")
}

func ArticleListApi(ctx iris.Context) {
	currentPage := ctx.URLParamIntDefault("page", 1)
	pageSize := ctx.URLParamIntDefault("limit", 20)

	articleList, total, err := provider.GetArticleList(currentPage, pageSize)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code":  config.StatusOK,
		"msg":   "",
		"data":  articleList,
		"count": total,
	})
}

func ArticleDeleteApi(ctx iris.Context) {
	var req request.Article
	if err := ctx.ReadForm(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	article, err := provider.GetArticleById(req.ID)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	err = article.Delete()
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "删除成功",
	})
}

func ArticleSourceListApi(ctx iris.Context) {
	currentPage := ctx.URLParamIntDefault("page", 1)
	pageSize := ctx.URLParamIntDefault("limit", 20)

	sourceList, total, err := provider.GetArticleSourceList(currentPage, pageSize)
	list := response.FormatArticleSourceList(sourceList)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code":  config.StatusOK,
		"msg":   "",
		"data":  list,
		"count": total,
	})
}

func ArticleSourceDeleteApi(ctx iris.Context) {
	var req request.ArticleSource
	if err := ctx.ReadForm(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	source, err := provider.GetArticleSourceById(req.ID)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	err = source.Delete()
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "删除成功",
	})
}

func ArticleSourceSaveApi(ctx iris.Context) {
	var req request.ArticleSource
	err := ctx.ReadJSON(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	var source *model.ArticleSource
	if req.ID > 0 {
		source, err = provider.GetArticleSourceById(req.ID)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
	} else {
		source, err = provider.GetArticleSourceByUrl(req.Url)
		if err == nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  "该数据源已存在，不用重复添加",
			})
			return
		}
		source = &model.ArticleSource{}
		source.Url = req.Url
	}

	if req.Url != "" {
		if strings.HasPrefix(req.Url, "http") == false {
			req.Url = "http://"+req.Url
		}
		source.Url = req.Url
	}
	source.UrlType 			= req.UrlType
	if source.UrlType == 2{
		source.IsMonitor = 0
	}
	source.IsMonitor  = req.IsMonitor
	articleSourceAttr := &model.ArticleSourceAttr{}

	fieldsData, _ := json.Marshal(req.Rule)
	articleSourceAttr.Rule = string(fieldsData[:])
	source.Attr = articleSourceAttr

	err = source.Save()
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	//添加完，马上抓取
	core.GetArticleLinks(source)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "添加/修改成功",
		"data": source,
	})
}

func ArticlePublishApi(ctx iris.Context) {
	var req request.Article
	if err := ctx.ReadForm(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	article, err := provider.GetArticleById(req.ID)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	core.AutoPublish(article)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "推送成功",
	})
}

func ArticleCatchApi(ctx iris.Context) {
	var req request.Article
	if err := ctx.ReadForm(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	article, err := provider.GetArticleById(req.ID)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	go core.GetArticleDetail(article)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "抓取任务已执行",
	})
}

func ArticleSourceCatchApi(ctx iris.Context) {
	var req request.ArticleSource
	if err := ctx.ReadForm(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	source, err := provider.GetArticleSourceById(req.ID)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	go core.GetArticleLinks(source)

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "抓取任务执行",
	})
}