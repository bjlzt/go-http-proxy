package main

import "fmt"
import "net/http"
import "strings"
import "io/ioutil"
import "time"

//发送请求的客户端
//var client = http.DefaultClient
var timeout = time.Duration(5 * time.Second)
var client = http.Client{
	Timeout: timeout,
}

//要跟踪的白名单
type whiteList struct {
	hosts []string
}

var wl = &whiteList{hosts: []string{"snssdk.com", "umeng.co"}}

//是否在白名单中
func (wl *whiteList) In(host string) (in bool) {
	for _, v := range wl.hosts {
		if v == "*" {
			return true
		} else if strings.Contains(host, v) {
			in = true
			break
		}
	}
	return
}

type MyMux struct{}

func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var scheme string = strings.ToLower(r.URL.Scheme)
	if scheme == "https" {
		fmt.Fprintf(w, "https not supported")
		return
	}
	method := strings.ToUpper(r.Method)
	log("proxy for=" + scheme + "://" + r.URL.Host + r.URL.Path + "?" + r.URL.RawQuery + ",method=" + method)

	switch {
	case method == "POST":
		post(w, r)
	case method == "GET":
		get(w, r)
	case method == "CONNECT":
		connect(w, r)
	default:
		fmt.Fprintf(w, "not support:"+method)
	}
	return
}

//发送请求
func do(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Accept-Encoding", "") //设为空，返回值将不会压缩

	resp, e := client.Do(r)
	if e != nil {
		log(e.Error())
		return
	}
	defer resp.Body.Close()
	tmp, er := ioutil.ReadAll(resp.Body)
	if er != nil {
		log(er.Error())
		return
	}
	body := string(tmp)
	fmt.Fprintf(w, body)
	parse(body, resp)
}

//METHOD == 'CONNECT'
func connect(w http.ResponseWriter, r *http.Request) {
	nr := &http.Request{
		Method:     "CONNECT",
		URL:        r.URL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     r.Header,
		Body:       nil,
		Host:       r.Host,
	}
	do(w, nr)
}

//METHOD == 'GET'
func get(w http.ResponseWriter, r *http.Request) {
	nr := &http.Request{
		Method:     "GET",
		URL:        r.URL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     r.Header,
		Body:       nil,
		Host:       r.Host,
	}
	do(w, nr)
}

//METHOD == 'POST'
func post(w http.ResponseWriter, r *http.Request) {
	nr := &http.Request{
		Method:     "POST",
		URL:        r.URL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     r.Header,
		Body:       r.Body,
		Host:       r.Host,
	}
	do(w, nr)
}

//打印白名单中的请求：request、response
func parse(body string, resp *http.Response) {
	host := strings.ToLower(resp.Request.Host)
	if !wl.In(host) {
		return
	}
	rstr := ""
	for k, v := range resp.Request.Header {
		rstr += k + ":" + strings.Join(v, "|") + "\n"
	}
	str := ""
	for k, v := range resp.Header {
		str += k + ":" + strings.Join(v, "|") + "\n"
	}
	log("*********************************************")
	log("Request Headers:")
	log("*********************************************")
	log(rstr)
	log("*********************************************")
	log("Response Headers:")
	log("*********************************************")
	log(str)
	format := resp.Header.Get("Content-Type")
	if format == "application/json" {

	}
	log(body)
	log("##########################################################################")
}

//记录日志
func log(log string) {
	fmt.Println(log)
}

func main() {
	mux := &MyMux{}
	http.ListenAndServe(":9090", mux)
}
