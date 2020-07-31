package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codepository/user/model"
)

// AddNewLabel 添加新标签
func AddNewLabel(c *model.Container) error {
	//
	// 检查参数
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New("参数类型{'header':{'token':''},'body':{'data':[{'name':'','type':'','describe':''},{}]}}")
	}
	// 检查类型

	// 保存
	var strbuffer strings.Builder
	for i, v := range c.Body.Data {
		json := v.(map[string]interface{})
		l := &model.WeixinOauserTag{}
		err := l.FromMap(json)
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第%d条失败:%s", i+1, err.Error()))
			continue
		}
		err = l.SaveOrUpdate()
		if err != nil {
			strbuffer.WriteString(fmt.Sprintf("第%d条失败:%s", i+1, err.Error()))
			continue
		}
	}
	c.Body.Data = nil

	if strbuffer.Len() != 0 {

		return errors.New(strbuffer.String())
	}
	c.Header.Msg = "添加成功"
	return nil
}

// FindTagidsByTagName 根据标签名称返回标称id
func FindTagidsByTagName(tagnames []interface{}) ([]int, error) {
	tags, err := model.FindAllTags("tagName in (?)", tagnames)
	if err != nil {
		return nil, err
	}
	var tagarr []int
	for _, tag := range tags {
		tagarr = append(tagarr, tag.ID)
	}
	return tagarr, nil
}
