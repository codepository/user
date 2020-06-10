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
