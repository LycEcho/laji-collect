package provider

import (
	"lajiCollect/model"
	"lajiCollect/services"
)

func GetArticleSourceList(currentPage int, pageSize int) ([]model.ArticleSource, int, error) {
	var sources []model.ArticleSource
	offset := (currentPage - 1) * pageSize
	var total int

	builder := services.DB.Model(model.ArticleSource{}).Order("id desc")
	if err := builder.Count(&total).Limit(pageSize).Offset(offset).Find(&sources).Error; err != nil {
		return nil, 0, err
	}
	return sources, total, nil
}

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

func GetArticleSourceById(id int) (*model.ArticleSource, error) {
	var source model.ArticleSource
	if err := services.DB.Model(model.ArticleSource{}).Where("`id` = ?", id).First(&source).Error; err != nil {
		return nil, err
	}

	return &source, nil
}

func GetArticleSourceByUrl(uri string) (*model.ArticleSource, error) {
	var source model.ArticleSource
	if err := services.DB.Model(model.ArticleSource{}).Where("`url` = ?", uri).First(&source).Error; err != nil {
		return nil, err
	}

	return &source, nil
}
