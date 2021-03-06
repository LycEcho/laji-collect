package library

import "strings"

func HasPrefix(need string, needArray []string) bool {
	for _, v := range needArray {
		if strings.HasPrefix(need, v) {
			return true
		}
	}

	return false
}

func HasSuffix(need string, needArray []string) bool {
	for _, v := range needArray {
		if strings.HasSuffix(need, v) {
			return true
		}
	}

	return false
}

//判断数组是否存在字符串
func HasContain(need string, needArray []string) bool {
	for _, v := range needArray {
		if strings.Contains(need, v) {
			return true
		}
	}

	return false
}
