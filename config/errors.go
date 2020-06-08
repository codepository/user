package config

type errorStatus int

const (
	// CheckNewUserErr 检查新用户
	CheckNewUserErr errorStatus = 0 + iota
	// AddLabelErr 添加标签错误
	AddLabelErr
	// UpdateUserMapErr 更新用户信息缓存错误
	UpdateUserMapErr
)

// Errors Errors
var Errors = []string{"检查新用户错误", "添加标签错误", "更新用户信息缓存错误"}
