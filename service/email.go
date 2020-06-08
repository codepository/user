package service

import (
	"strconv"

	"github.com/codepository/user/config"
	"gopkg.in/gomail.v2"
)

// SendMail 发送邮件
// subject 标题
// body 内容
func SendMail(mailTo []string, from, subject string, body string) error {
	config := config.Config
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	port, _ := strconv.Atoi(config.EmailPort)
	d := gomail.NewDialer(config.EmailHost, port, config.EmailUser, config.EmailPass)
	err := d.DialAndSend(m)
	return err
}
