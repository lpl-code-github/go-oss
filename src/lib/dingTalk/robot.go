package dingTalk

import (
	"fmt"
	"github.com/CodyGuo/dingtalk"
	"os"
	"oss/src/lib/myLog"
)

func RobotSend(markdownContent string) int {
	webHook := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", os.Getenv("DINGTALK_TOKEN"))
	secret := os.Getenv("DINGTALK_SECRET")
	dt := dingtalk.New(webHook, dingtalk.WithSecret(secret))
	// markdown类型
	markdownTitle := "markdown"
	if err := dt.RobotSendMarkdown(markdownTitle, markdownContent); err != nil {
		myLog.Error.Println(err)
	}
	return printResult(dt)
}

func printResult(dt *dingtalk.DingTalk) int {
	response, err := dt.GetResponse()
	if err != nil {
		myLog.Error.Println(err)
	}

	return response.StatusCode
}
