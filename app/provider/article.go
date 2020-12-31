package provider

import (
	"lajiCollect/model"
	"lajiCollect/services"
)
//获得采集源列表
func GetArticleSourceList(currentPage int, pageSize int) ([]*model.ArticleSource, int, error) {
	sources := []*model.ArticleSource{}
	offset := (currentPage - 1) * pageSize
	var total int

	builder := services.DB.Model(model.ArticleSource{}).Order("id desc")
	if err := builder.Count(&total).Limit(pageSize).Offset(offset).Find(&sources).Error; err != nil {
		return nil, 0, err
	}
	for _,v := range sources{
		GetArticleSourceInfo(v)
	}
	return sources, total, nil
}

//获取采集源的附表数据
func GetArticleSourceInfo(articleSource *model.ArticleSource){
	if articleSource.Attr == nil {
		attr := &model.ArticleSourceAttr{}
		services.DB.Model(model.ArticleSourceAttr{}).Where("source_id = ?",articleSource.Id).Take(&attr)
		articleSource.Attr = attr
	}
}
//获得文章列表
func GetArticleList(currentPage int, pageSize int) ([]model.Article, int, error) {
	var articles []model.Article
	offset := (currentPage - 1) * pageSize
	var total int

	builder := services.DB.Model(model.Article{}).Order("id desc")
	if err := builder.Count(&total).Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	if len(articles) > 0 {
		for i, v := range articles {
			var articleData model.ArticleData
			if err := services.DB.Model(model.ArticleData{}).Where("`id` = ?", v.Id).First(&articleData).Error; err == nil {
				articles[i].Content = articleData.Content
			}
		}
	}
	return articles, total, nil
}
//获得文章 根据Id
func GetArticleById(id int) (*model.Article, error) {
	var article model.Article
	if err := services.DB.Model(model.Article{}).Where("`id` = ?", id).First(&article).Error; err != nil {
		return nil, err
	}
	var articleData model.ArticleData
	if err := services.DB.Model(model.ArticleData{}).Where("`id` = ?", id).First(&articleData).Error; err != nil {
		return nil, err
	}
	article.Content = articleData.Content

	return &article, nil
}
//获得采集源 根据Id
func GetArticleSourceById(id int) (*model.ArticleSource, error) {
	source := &model.ArticleSource{}
	if err := services.DB.Where("`id` = ?", id).First(source).Error; err != nil {
		return nil, err
	}
	GetArticleSourceInfo(source)
	return source, nil
}
//获得采集源 根据Url
func GetArticleSourceByUrl(uri string) (*model.ArticleSource, error) {
	source := &model.ArticleSource{}
	if err := services.DB.Model(model.ArticleSource{}).Where("`url` = ?", uri).First(source).Error; err != nil {
		return nil, err
	}
	GetArticleSourceInfo(source)
	return source, nil
}
