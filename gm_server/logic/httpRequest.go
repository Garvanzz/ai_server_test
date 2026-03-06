package logic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// 请求
func HttpRequest(jsonData []byte, msgInterface string) (error, string) {
	// 创建自定义客户端
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:9505/api/"+msgInterface, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err, ""
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer your-token-here")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err, ""
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return err, ""
	}

	fmt.Println("Status Code:", resp.Status)
	fmt.Println("Response:", string(body))
	return nil, string(body)
}
