package request

type ArticleSource struct {
	ID         			 int    					`form:"id"`
	Url        			 string 					`form:"url" validate:"required"`
	ErrorTimes 			 int    					`form:"error_times"`
	UrlType    			 int    					`form:"url_type"`
	Rule    			 ArticleSourceAttrRule  	`form:"rule"`
}
type ArticleSourceAttrRule struct {
	UrlOnlySelf int8 `form:"url_only_self";json:"url_only_self"` //是否过滤非本站点的链接
}
type Article struct {
	ID int `form:"id"`
}