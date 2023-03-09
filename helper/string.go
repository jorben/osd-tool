package helper

import (
	"math"
	"strconv"
	"strings"
)

// HideSecret 隐藏字符串的中间字符
func HideSecret(secret string, count uint32) string {
	length := len(secret)
	if 0 == length {
		return ""
	}
	mask := strings.Repeat("*", int(count))
	if length <= int(count) {
		return mask[0:length]
	}

	prefix := math.Ceil(float64(length-int(count)) / 2)
	suffix := math.Floor(float64(length-int(count)) / 2)

	return secret[0:int(prefix)] + mask + secret[length-int(suffix):]
}

// CompareVersion 版本号对比
// 两版本相同时返回0，
// version1 > version2时返回1，
// version1 < version2时返回-1
func CompareVersion(version1, version2 string) int {
	// 将版本号按"."分割
	v1 := strings.Split(strings.TrimLeft(strings.ToLower(version1), "v"), ".")
	v2 := strings.Split(strings.TrimLeft(strings.ToLower(version2), "v"), ".")
	// 判断版本号的长度
	len1, len2 := len(v1), len(v2)
	vlen := len1
	if len1 < len2 {
		vlen = len1
	} else {
		vlen = len2
	}
	// 遍历分割后的版本号字符串，从左到右依次比较
	for i := 0; i < vlen; i++ {
		// 将字符串转换为整数，再比较大小
		num1, _ := strconv.Atoi(v1[i])
		num2, _ := strconv.Atoi(v2[i])
		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}
	// 如果仍相等，再判断长度
	if len1 > len2 {
		return 1
	} else if len1 < len2 {
		return -1
	}
	return 0
}
