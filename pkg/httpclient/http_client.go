
package httpclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cyruslo/library/pkg/beego/httplib"
	LOGGER "github.com/cyruslo/util/logger"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	clientCACert []byte
)
///////////////////////////////////
//http_client采用beego框架的httplib库,需要新增方法参考:https://beego.me/docs/module/httplib.md

//get请求str
func HttpGetStr(host string, request string, content string) string {
	url := fmt.Sprintf("%s/%s?%s", host, request, content)
	
	log.Printf("url:%s.", url)
	
	rep := httplib.Get(url)	
	
	str, err := rep.String()
	
	if err != nil {
		log.Printf("httplib err")
		
		fmt.Println(err)
		return ""
	}

//	fmt.Println(str)
	return str
}

func HttpsGetStr(host, request, content string) string {
	pool := x509.NewCertPool()
 
    pool.AppendCertsFromPEM(clientCACert)
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{RootCAs: pool},
    }
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("%s/%s?%s", host, request, content)
	log.Printf("url:%s.", url)
	
    resp, err := client.Get(url)
    if err != nil {
        fmt.Println("Get error:", err)
        return ""
    }
    defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	res := string(body)
	//fmt.Println(res)
	return res
}
/*
//get请求json
func HttpGetJson(host string, request string, content string) interface{} {
	url := fmt.Sprintf("%s/%s?%s", host, request, content)

	rep := httplib.Get(url)
	str, err := rep.String()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println(str)
	var result interface{}
	return rep.ToJSON(&result)
}*/

//post请求string
func HttpPostStr(host string, request string, content string, values map[string]string) string {
	url := fmt.Sprintf("%s/%s?%s", host, request, content)

	rep := httplib.Post(url)
	for key := range values {
		rep.Param(key, values[key])
	}
	
	str, err := rep.String()
	if err != nil {
		fmt.Println(err)
		return ""
	}	

//	fmt.Println(str)
	return str
}

//post请求xml		
func HttpPostXml(url string, output []byte)  {

    var jsonStr = []byte(output)
 
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    // req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
		fmt.Println("client.Do err:%v.", err)
		return
    }
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    LOGGER.Info("response Body:", string(body))
}

func OnRead(caCertPath string) {

    caCrt, err := ioutil.ReadFile(caCertPath)
    if err != nil {
        fmt.Println("ReadFile err:", err)
        return
	}
	
	clientCACert = caCrt
}