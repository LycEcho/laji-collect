package constant
const(
	DbArticleUrlTypeList = 1 				//万能列表
	DbArticleUrlTypeDetail = 2 				//万能详情
	DbArticleUrlTypeWordpressRss = 3 		//WordpressRss类型

	DbArticleStatusWait = 0  //待采集
	DbArticleStatusPass = 1  //有效数据
	DbArticleStatusIng = 2  //采集中
	DbArticleStatusFial = 3  //无效数据


	DbArticleStatusReleaseUn = 0  //未发布
	DbArticleStatusReleaseEd = 1  //已发布
	DbArticleStatusReleaseIng = 2  //发布中
)