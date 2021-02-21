package httpreq

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jacktrane/gocomponent/format"
)

var (
	ConnectTimeOut   = 10 * time.Second
	ReadWriteTimeOut = 30 * time.Second
)

//通用http请求
type HttpRequest struct {
	url      string
	req      *http.Request
	params   map[string]interface{}
	resp     *http.Response
	respBody []byte

	//连接超时时间，秒
	connectTimeOut time.Duration
	//读写超时时间，秒
	readWriteTimeout time.Duration
	//连接后等待响应的时长 秒
	responseTimeout time.Duration
	//用户UserAgent
	userAgent string

	isDebug bool

	TimeOut int64
}

//实例化一个httpRequest
//@param rawUrl string
//@param method string 请求的方法，GET/POST
func NewRequest() *HttpRequest {
	httpR := &HttpRequest{}
	return httpR.Init()
}

//构建req
//@param rawUrl string
//@param method string 请求的方法，GET/POST
func (h *HttpRequest) Init() *HttpRequest {
	req := &http.Request{
		//Method:     method,
		//URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		//Host:       u.Host,
	}
	//h.url = rawUrl
	h.req = req
	h.params = make(map[string]interface{})
	h.connectTimeOut = ConnectTimeOut
	h.readWriteTimeout = ReadWriteTimeOut
	h.resp = &http.Response{}
	h.TimeOut = time.Now().UnixNano()
	return h
}

func (h *HttpRequest) GetUrl() string {
	return h.req.URL.String()
}

func (h *HttpRequest) SetMethod(method string) *HttpRequest {
	h.req.Method = method
	return h
}

func (h *HttpRequest) SetBasicAuth(username, password string) *HttpRequest {
	h.req.SetBasicAuth(username, password)
	return h
}

func (h *HttpRequest) SetUrl(url string) *HttpRequest {
	h.url = url
	return h
}

//设置连接超时
func (h *HttpRequest) SetConnectTimeOut(c time.Duration) *HttpRequest {
	h.connectTimeOut = c
	return h
}

//设置读超时
func (h *HttpRequest) SetReadWriteTimeout(r time.Duration) *HttpRequest {
	h.readWriteTimeout = r
	return h
}

//设置UserAgent
func (b *HttpRequest) SetUserAgent(userAgent string) *HttpRequest {
	b.userAgent = userAgent
	return b
}

//是否打印请求日志
func (b *HttpRequest) Debug(isDebug bool) *HttpRequest {
	b.isDebug = isDebug
	return b
}

//构建url
func (h *HttpRequest) buildURL() {
	if h.req.Method == "GET" && len(h.params) > 0 {
		if strings.Contains(h.url, "?") {
			h.url = h.url + "&" + buildQuery(h.params)
		} else {
			h.url = h.url + "?" + buildQuery(h.params)
		}
		return
	}
	if h.req.Method == "POST" || h.req.Method == "PUT" {
		// if len(h.params) > 0 {
		// 	h.Header("Content-Type", "application/x-www-form-urlencoded")
		if h.req.Header.Get("Content-Type") == "application/json" {
			dataEncode, _ := json.Marshal(h.params)
			h.Body(dataEncode)
		} else {
			h.Body(buildQuery(h.params))
		}
		// }
	}
}

//在body设置请求数据
func (h *HttpRequest) Body(data interface{}) *HttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		h.req.Body = ioutil.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		h.req.Body = ioutil.NopCloser(bf)
		h.req.ContentLength = int64(len(t))
	}
	return h
}

//设置多个参数
func (h *HttpRequest) Params(p map[string]interface{}) *HttpRequest {
	for k, v := range p {
		h.params[k] = v
	}
	return h
}

//设置单个参数
func (h *HttpRequest) Param(key string, value interface{}) *HttpRequest {
	h.params[key] = value
	return h
}

// 格式化参数成json

//设置Header
func (h *HttpRequest) Header(key, value string) *HttpRequest {
	h.req.Header.Set(key, value)
	return h
}

//构建参数
func buildQuery(params map[string]interface{}) (paramBody string) {

	var buf bytes.Buffer
	for k, v := range params {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(format.ToString(v)))
		buf.WriteByte('&')
	}
	paramBody = buf.String()
	if len(paramBody) > 0 {
		paramBody = paramBody[0 : len(paramBody)-1]
	}
	return paramBody
}

//获取返回值
func (h *HttpRequest) Response() (*http.Response, error) {
	if h.resp.StatusCode != 0 {
		return h.resp, nil
	}
	resp, err := h.doRequest()
	h.TimeOut = (time.Now().UnixNano() - h.TimeOut) / 1000000
	if err != nil {
		return nil, err
	}
	h.resp = resp

	return resp, nil
}

//执行请求
func (h *HttpRequest) doRequest() (*http.Response, error) {
	h.buildURL()
	u, err := url.Parse(h.url)
	if err != nil {
		return nil, err
	}
	h.req.URL = u
	trans := &http.Transport{
		Dial: h.timeoutDialer(h.connectTimeOut, h.readWriteTimeout),
	}
	h.Header("Connection", "close").Header("Accept-Charset", "utf-8").Header("Cache-Control", "max-age=0")
	if h.userAgent != "" && h.req.Header.Get("User-Agent") == "" {
		h.Header("User-Agent", h.userAgent)
	}
	client := &http.Client{
		Transport: trans,
	}

	return client.Do(h.req)
}

//设置timeout
func (h *HttpRequest) timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

//获取bytes
func (h *HttpRequest) Bytes() ([]byte, error) {
	if h.respBody != nil {
		return h.respBody, nil
	}
	resp, err := h.Response()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	h.respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return h.respBody, nil
}

//获取string
func (h *HttpRequest) String() (string, error) {
	data, err := h.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func BuildQuery(req map[string]interface{}) string {
	var buf bytes.Buffer
	for k, v := range req {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(format.ToString(v)))
	}
	return buf.String()
}
