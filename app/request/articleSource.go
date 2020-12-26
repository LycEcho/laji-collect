package request

type ArticleSource struct {
	ID         			 int    `form:"id"`
	Url        			 string `form:"url" validate:"required"`
	ErrorTimes 			 int    `form:"error_times"`
	UrlType    			 int    `form:"url_type"`
	UrlRuleOnlyMyself    uint8  `form:"url_rule_only_myself"`
}

type Article struct {
	ID int `form:"id"`
}