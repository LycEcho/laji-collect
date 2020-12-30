package response

import (
	"encoding/json"
	"lajiCollect/model"
)

type ArticleSourceList struct {
	ID          int    `form:"id"`
	Url         string `form:"url" validate:"required"`
	ErrorTimes  int    `form:"error_times"`
	UrlType     int    `form:"url_type"`
	UrlOnlySelf int8   `json:"rule.url_only_self"` //是否过滤非本站点的链接
}

func FormatArticleSourceList(source []model.ArticleSource) []*ArticleSourceList {
	var list []*ArticleSourceList
	for _,v := range source{
		fieldsData, _ :=  json.Marshal(v)
		resp := &ArticleSourceList{}
		if err := json.Unmarshal([]byte(fieldsData), resp); err != nil{
			resp = &ArticleSourceList{}
		}
		list = append(list,resp)
	}
	return list
}

