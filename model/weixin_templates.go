package model

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/mumushuiding/util"
)

// WeixinTemplatesTable 微信流程模板
var WeixinTemplatesTable = "weixin_templates"

// WeixinTemplates 流程模板
type WeixinTemplates struct {
	ID           int    `gorm:"primary_key" json:"id,omitempty"`
	TemplateID   string `gorm:"size:50;column:templateId" json:"templateId"`
	TemplateName string `gorm:"column:templateName" json:"templateName"`
	// 0-提交申请时，1-审批通过时，3-提交和审批者抄送
	NotifyAttr   int    `gorm:"column:notifyAttr" json:"notifyAttr"`
	TemplateData string `gorm:"column:templateData" json:"templateData"`
	Appid        int    `json:"appid"`
	Appname      string `json:"appname"`
}

// GetTemplateData 获取流程数据
func (w *WeixinTemplates) GetTemplateData() (*Node, error) {
	w.TemplateData = strings.ReplaceAll(w.TemplateData, "\r", "")
	w.TemplateData = strings.ReplaceAll(w.TemplateData, "\n", "")
	n := &Node{}
	err := util.Str2Struct(w.TemplateData, n)
	if err != nil {
		return nil, fmt.Errorf("流程字符串转对象,err:%s", err.Error())
	}
	return n, nil
}

// FindByTemplateID FindByTemplateID
func (w *WeixinTemplates) FindByTemplateID(templateID string) error {
	return wxdb.Table(WeixinTemplatesTable).Where("templateId=?", templateID).Find(w).Error
}

// GenerateThirdNo 生成流水号
func (w *WeixinTemplates) GenerateThirdNo(userID string) string {
	p := sha256.Sum256(append([]byte(userID), []byte(time.Now().String())...))
	return string(util.Base58Encode(p[:10]))
}
