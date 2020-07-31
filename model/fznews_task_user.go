package model

import (
	"fmt"
	"strconv"

	"github.com/mumushuiding/util"
)

// FznewsTaskUserTableName FznewsTaskUserTableName
var FznewsTaskUserTableName = "fznews_task_user"

// FznewsTaskUser 用户待完成任务及完成情况
type FznewsTaskUser struct {
	Model
	// 用户id
	Userid   int    `json:"userid,omitempty"`
	Username string `json:"username,omitempty"`
	// 用户需要完成的任务
	Task string `json:"task,omitempty"`
	// 完成情况 0-未完成，1-完成
	Complete int `json:"complete,omitempty"`
	// 用户标签
	Role string `json:"role,omitempty"`
	// 任务开始时间
	Start string `json:"start,omitempty"`
	// 任务结束时间
	End string `json:"end,omitempty"`
	// 完成时间
	FinishTime string `json:"finish_time,omitempty"`
}

// ToString ToString
func (f *FznewsTaskUser) ToString() string {
	s, _ := util.ToJSONStr(f)
	return s
}

// Save 插入新的字段
func (f *FznewsTaskUser) Save() error {
	return db.Table(FznewsTaskUserTableName).Create(f).Error
}

// FindTaskCompleteRate 任务完成率
func FindTaskCompleteRate(query interface{}, values ...interface{}) (float64, error) {
	var total int
	var complete int
	err := db.Table(FznewsTaskUserTableName).Where(query, values...).Count(&total).Where("complete=?", 1).Count(&complete).Error
	if err != nil {
		return 0, err
	}
	// err = db.Table(FznewsTaskUserTableName).Where(query, values...).Where("complete=?", 1).Count(&complete).Error
	// if err != nil {
	// 	return 0, err
	// }
	// log.Printf("总数：%d,完成数:%d", total, complete)
	r := fmt.Sprintf("%.2f", float64(complete)/float64(total))
	x, _ := strconv.ParseFloat(r, 64)
	return x, nil
}

// FindUsersUnCompleteTask FindUsersUnCompleteTask
func FindUsersUnCompleteTask(limit, offset int, query interface{}, values ...interface{}) ([]*FznewsTaskUser, error) {
	var datas []*FznewsTaskUser
	err := db.Table(FznewsTaskUserTableName).
		Select("userid,username").
		Where(query, values).Where("complete=?", 0).Limit(limit).Offset(offset).Find(&datas).Error
	return datas, err
}
