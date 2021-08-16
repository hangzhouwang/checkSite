//Author: 西瓜哥
//Github: https://github.com/siaoynli
//LastEditors: 西瓜哥
//Date: 2021-08-12 15:40:47
//LastEditTime: 2021-08-16 11:01:51
//Description:
//Copyright:  Copyright (c)  2021 by http://www.hangzhou.com.cn,All Rights Reserved.

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"
	"gopkg.in/gomail.v2"
)

func SendMail(config *toml.Tree, url string, errMsg string) error {

	host := config.Get("mailserver.host").(string)
	_port := config.Get("mailserver.port").(string)
	user := config.Get("mailserver.user").(string)
	pass := config.Get("mailserver.pass").(string)

	mailConn := map[string]string{
		"user": user,
		"pass": pass,
		"host": host,
		"port": _port,
	}

	port, _ := strconv.Atoi(mailConn["port"])

	mailto := config.Get("mailto.users").(string)

	users := strings.Split(mailto, "|")

	mailTo := make([]string, 0)

	mailTo = append(mailTo, users...)

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(mailConn["user"], config.Get("server").(string)))
	m.SetHeader("To", mailTo...)                                        //发送给多个用户
	m.SetHeader("Subject", config.Get("mailto.subject").(string))       //设置邮件主题
	m.SetBody("text/html", fmt.Sprintf("%s网址检测失败,错误信息%s", url, errMsg)) //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err

}

func main() {

	str, _ := os.Getwd()
	config, err := toml.LoadFile(fmt.Sprintf("%s/config.toml", str)) //加载toml文件
	if err != nil {
		panic(err.Error())
	}
	urls := config.Get("urls").(string)
	urlsArr := strings.Split(urls, "|")
	urlArr := make([]string, 0)
	urlArr = append(urlArr, urlsArr...)
	wg := sync.WaitGroup{}
	for _, v := range urlArr {
		wg.Add(1)
		fmt.Println("url:" + v)
		go func(v string) {
			defer wg.Done()
			res, err := http.Get(v)
			if err != nil {
				fmt.Printf("网址:%s,错误:%s\n", v, err.Error())
				err = SendMail(config, v, err.Error())
				if err != nil {
					fmt.Println("发送邮件失败，请检查代码")
				}
				fmt.Println("邮件发送成功")
				return
			}
			if res.StatusCode != 200 {
				fmt.Printf("网址:%s,状态码:%d\n", v, res.StatusCode)
				err = SendMail(config, v, "状态码不是200")
				if err != nil {
					fmt.Println("发送邮件失败，请检查代码")
					return
				}
				fmt.Println("邮件发送成功")
			}

		}(v)
	}

	wg.Wait()

	fmt.Println("状态检测完毕")

}
