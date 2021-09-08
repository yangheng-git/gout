package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/guonaihong/gout"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// response拦截器修改示例
type demoResponseMiddler struct{}

func (d *demoResponseMiddler) ModifyResponse(response *http.Response) error {
	// 修改responseBody。 因为返回值大概率会有 { code, data,msg} 等字段,希望进行统一处理
	//这里想验证code. 如果不对就返回error。 对的话将 data中的内容写入body,这样后面BindJson的时候就可以直接处理业务了
	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	obj := make(map[string]interface{})
	err = json.Unmarshal(all, &obj)
	if err != nil {
		return err
	}
	code := obj["code"]
	msg := obj["msg"]
	data := obj["data"]

	// Go中json中的数字经过反序列化会成为float64类型
	if float64(200) != code {
		return errors.New(fmt.Sprintf("请求失败, code %d msg %s", code, msg))
	} else {
		byt, _ := json.Marshal(&data)
		response.Body = ioutil.NopCloser(bytes.NewReader(byt))
		return nil
	}
}
func demoResponse() *demoResponseMiddler {
	return &demoResponseMiddler{}
}

func main() {
	go server()                        //等会起测试服务
	time.Sleep(time.Millisecond * 500) //用时间做个等待同步
	responseUseExample()
}

func responseUseExample() {
	//成功请求
	successRes := new(map[string]interface{})
	err := gout.GET(":8080/success").ResponseUse(demoResponse()).BindJSON(&successRes).Do()
	log.Printf("success请求  -->   响应 %s  \n  err  %s \n ", successRes, err)

	//fail请求
	failRes := new(map[string]interface{})
	err = gout.GET(":8080/fail").ResponseUse(demoResponse()).BindJSON(&failRes).Do()
	log.Printf("fail请求  -->   响应 %s  \n  err  %s \n ", failRes, err)
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
type Item struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func server() {
	router := gin.New()
	router.GET("/success", func(c *gin.Context) {
		c.JSON(200, Result{200, "请求成功了", Item{"001", "张三"}})
	})
	router.GET("/fail", func(c *gin.Context) {
		c.JSON(200, Result{500, "查询数据库出错了", nil})
	})
	router.Run()
}
