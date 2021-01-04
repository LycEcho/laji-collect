package response

import (
	"encoding/json"
	"lajiCollect/model"
)

type ArticleSourceList struct {
	ID          int    `json:"id"`
	Url         string `json:"url" validate:"required"`
	ErrorTimes  int    `json:"errorTimes"`
	UrlType     int    `json:"urlType"`
	IsMonitor   int8   `json:"isMonitor"`
	UrlOnlySelf int8   `json:"urlOnlySelf"` //是否过滤非本站点的链接
	OnlyText 	int8   `json:"onlyText"` //是否过滤非本站点的链接
}

func FormatArticleSourceList(source []*model.ArticleSource) []*ArticleSourceList {
	list := []*ArticleSourceList{}
	for _, v := range source {
		fieldsData, _ := json.Marshal(v)
		resp := &ArticleSourceList{}
		if err := json.Unmarshal([]byte(fieldsData), resp); err != nil {
			resp = &ArticleSourceList{}
		}

		rule,err := v.GetParseRule()
		if err == nil {
			//转化attr
			ruleData, _ := json.Marshal(rule)
			json.Unmarshal([]byte(ruleData), resp)
		}

		list = append(list, resp)
	}
	return list
}
