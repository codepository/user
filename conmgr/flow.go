package conmgr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codepository/user/model"
)

// StartFlowByToken 启动流程
func StartFlowByToken(c *model.Container) error {
	if len(c.Header.Token) == 0 || len(c.Body.Metrics) == 0 {
		return errors.New(`参数类型{"header:{"token":""},"body":{"metrics":"002d5df2a737dd36a2e78314b07d0bb1_1591669930"}},metrics为模板id`)
	}
	user, err := GetUserByToken(c.Header.Token)
	if err != nil {
		return err
	}
	labels, err := GetLabelNamesByToken(c.Header.Token)
	if err != nil {
		return err
	}
	var khz string
	for _, l := range labels {
		if strings.HasSuffix(l, "考核组成员") {
			khz = l
		}
	}
	if len(khz) == 0 {
		return errors.New("不是考核组成员不能填写考核表")
	}

	// c.Body.Data = append(c.Body.Data, params)

	// 查询流程
	templateID := c.Body.Metrics
	w := &model.WeixinTemplates{}
	err = w.FindByTemplateID(templateID)
	if err != nil {
		return err
	}
	// 流程数据
	n, err := w.GetTemplateData()
	if err != nil {
		return err
	}
	// 解析流程
	params := make(map[string]interface{})
	params["khzKey"] = khz
	params["userid"] = user.Userid
	params["zjKey"] = user.Level
	params["templateId"] = c.Body.Metrics
	params["isleader"] = float64(user.IsLeader)
	result, notify, err := n.Parse(params)
	if err != nil {
		return err
	}
	// 流水号
	thirdNo := w.GenerateThirdNo(user.Userid)
	// 存储执行流
	approvalData := model.Approvaldata{
		Errcode: 0,
		Errmsg:  "ok",
		Data: &model.ExecutionData{
			ThirdNo:        thirdNo,
			OpenTemplateID: w.TemplateID,
			OpenSpName:     w.TemplateName,
			OpenSpstatus:   1,
			ApplyTime:      time.Now().Unix(),
			ApplyUserID:    user.Userid,
			ApplyUsername:  user.Name,
			ApplyUserParty: user.Departmentname,
			ApplyUserImage: user.Avatar,
			ApprovalNodes:  result,
			NotifyNodes:    notify,
			Approverstep:   0,
		},
	}
	apd := &model.WeixinApprovaldata{
		Agentid:    w.Appid,
		ThirdNo:    w.GenerateThirdNo(user.Userid),
		Data:       approvalData.ToString(),
		Step:       0,
		Status:     1,
		NotifyAttr: 1,
	}

	err = apd.SaveOrUpdate()
	if err != nil {
		return fmt.Errorf("保存流程数据失败,%s", err.Error())
	}
	// 启动流程
	c.Body.Data = append(c.Body.Data, approvalData)
	return nil
}

// // StartFlow 启动流程
// func StartFlow(c *model.Container) error {
// 	// 检查参数
// 	if c.Body.Data == nil || len(c.Body.Data) == 0 {
// 		return errors.New(`参数类型必须是:{"body":{"data":[{"khzKey":"","zjKey":1,"userid":""}]}},khzKey表示用户所有考核组，zjKey表示用户的职级`)
// 	}
// 	params, yes := c.Body.Data[0].(map[string]interface{})
// 	if !yes {
// 		return errors.New(`参数类型必须是:{"body":{"data":[{"khzKey":"","zjKey":1,"userid":""}]}},khzKey表示用户所有考核组，zjKey表示用户的职级`)
// 	}

// 	return err

// }
