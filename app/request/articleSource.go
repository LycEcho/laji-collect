package request

type ArticleSource struct {
	ID         			 int    					`json:"id"`
	Url        			 string 					`json:"url"`
	UrlType    			 int    					`json:"urlType"`
	IsMonitor    		 int    					`json:"isMonitor"`
	Rule    			 ArticleSourceAttrRule 		`json:"rule"`
}
type ArticleSourceAttrRule struct {
	UrlOnlySelf 		int `json:"urlOnlySelf"` 	//是否过滤非本站点的链接
	OnlyText 			int `json:"onlyText"` 		//是否只保存文字
}
type Article struct {
	ID int `json:"id"`
}