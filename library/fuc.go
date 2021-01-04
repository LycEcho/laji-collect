package library

import (
	"fmt"
	"github.com/polaris1119/keyword"
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf8"
)


func InArray(need string, needArray []string) bool {
	for _, v := range needArray {
		if need == v {
			return true
		}
	}

	return false
}

//根据内容获取关键词
func GetKeywords(content string, num int) []string {
	var words []string
	length := 2
	keywords := keyword.Extractor.Extract(content, 1000)
	for _, v := range keywords {
		if utf8.RuneCountInString(v) >= length {
			words = append(words, v)
		}
	}

	if len(words) > num {
		return words[:num]
	}
	return words
}

//解析链接
func ParseLink(link string, baseUrl string) string {
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	if strings.Contains(link, "javascript") || strings.Contains(link, "void") || link == "#" || link == "./" || link == "../" || link == "../../" {
		return ""
	}

	link = replaceDot(link, baseUrl)

	return link
}

//相对链接转换绝对链接
func replaceDot(currUrl string, baseUrl string) string {
	if strings.HasPrefix(currUrl, "//") {
		currUrl = fmt.Sprintf("https:%s", currUrl)
	}
	urlInfo, err := url.Parse(currUrl)
	if err != nil {
		return ""
	}
	if urlInfo.Scheme != "" {
		return currUrl
	}
	baseInfo, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}

	u := baseInfo.Scheme + "://" + baseInfo.Host
	var path string
	if strings.Index(urlInfo.Path, "/") == 0 {
		path = urlInfo.Path
	} else {
		path = filepath.Dir(baseInfo.Path) + "/" + urlInfo.Path
	}

	rst := make([]string, 0)
	pathArr := strings.Split(path, "/")

	// 如果path是已/开头，那在rst加入一个空元素
	if pathArr[0] == "" {
		rst = append(rst, "")
	}
	for _, p := range pathArr {
		if p == ".." {
			if len(rst) > 0 {
				if rst[len(rst)-1] == ".." {
					rst = append(rst, "..")
				} else {
					rst = rst[:len(rst)-1]
				}
			}
		} else if p != "" && p != "." {
			rst = append(rst, p)
		}
	}
	return u + strings.Join(rst, "/")
}