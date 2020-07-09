package model

import (
	"strings"

	"github.com/mumushuiding/util"
)

// Container 参数和结果容器
type Container struct {
	Header CHeader `json:"header,omitemtpy"`
	Body   CBody   `json:"body,omitempty"`
}

// CHeader CHeader
type CHeader struct {
	Token string `json:"token,omitempty"`
	Msg   string `json:"msg,omitempty"`
}

// CBody 用于获取前台参数和返回结果
type CBody struct {
	Data       []interface{} `json:"data,omitempty"`
	Total      int           `json:"total,omitempty"`
	StartIndex int           `json:"start_index,omitempty"`
	MaxResults int           `json:"max_results,omitempty"`
	StartDate  string        `json:"start_date,omitempty"`
	EndDate    string        `json:"end_date,omitempty"`
	UserName   string        `json:"username,omitempty"`
	UserID     int           `json:"user_id,omitempty"`
	Method     string        `json:"method,omitempty"`
	Metrics    string        `json:"metrics,omitempty"`
	Fields     []string      `json:"fields,omitempty"`
}

// ToString ToString
func (c *Container) ToString() string {
	str, _ := util.ToJSONStr(c)
	return str
}

// FindUserinfoPaged 分页查询用户信息
func (c *Container) FindUserinfoPaged() error {
	if c.Body.MaxResults == 0 {
		c.Body.MaxResults = 50
	}
	if len(c.Body.Metrics) == 0 {
		c.Body.Metrics = "*"
	}
	var where strings.Builder
	if len(c.Body.UserName) != 0 {
		where.WriteString("and name like '%" + c.Body.UserName + "%'")
	}
	var result []*Userinfo

	var total int
	if where.Len() > 0 {
		err := wxdb.Table(UserinfoTabel).Select(c.Body.Metrics).
			Where(where.String()[3:]).
			Count(&total).
			Offset(c.Body.StartIndex).Limit(c.Body.MaxResults).
			Find(&result).Error
		if err != nil {
			return err
		}
	} else {
		err := wxdb.Table(UserinfoTabel).Select(c.Body.Metrics).
			Count(&total).
			Offset(c.Body.StartIndex).Limit(c.Body.MaxResults).
			Find(&result).Error
		if err != nil {
			return err
		}
	}
	c.Body.Data = append(c.Body.Data, result)
	c.Body.Total = total
	return nil
}
