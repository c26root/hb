package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hb/common"

	"github.com/gookit/color"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	wg                  sync.WaitGroup
	ch                  chan bool
	debug               bool
	displayProgress     bool
	displayResponseBdoy bool
	forceSSL            bool
	random              bool

	file             string
	f                *os.File
	reqHost          string
	method           string
	requestBody      string
	bodyFile         string
	path             string
	redirect         bool
	grepString       string
	filterString     string
	code             int
	proxies          string
	isReplace        bool
	result           []HttpInfo
	extractInfoReStr string
	extractInfoRe    *regexp.Regexp

	host       string
	port       string
	timeout    int
	threads    int
	outputFile string

	reqHeaders Headers
	bar        *pb.ProgressBar

	titleRe = regexp.MustCompile(`(?is)<title>\s?(.*?)\s?</title>`)
	headers = map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36",
	}

	successCount uint64
	errorCount   uint64
	matchCount   uint64
)

type Headers []string

type HttpInfo struct {
	StatusCode    int    `json:"statusCode"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	Server        string `json:"server"`
	ContentLength string `json:"length"`
	ContentType   string `json:"type"`
	XPoweredBy    string `json:xPoweredBy`
	ExtraInfo     string `json:"extraInfo"`
}

func (i *Headers) String() string {
	return strings.Join(*i, ", ")
}

func (i *Headers) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *HttpInfo) String() string {
	return strconv.Itoa(i.StatusCode) + i.URL + i.Title + i.Server + i.ContentLength + i.ContentType + i.XPoweredBy
}

func validMethod(method string) bool {
	methods := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func determineEncoding(r *bufio.Reader) encoding.Encoding {
	b, err := r.Peek(1024)
	if err != nil {
		// log.Error("get code error")
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(b, "")
	return e
}

func getProxyURL(proxyStr string) *url.URL {
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		log.Println(err)
	}
	return proxyURL
}

func init() {
	flag.StringVar(&host, "host", "", "host or host range. 127.0.0.1 | 192.168.1.1/24 | 192.168.1.1-5")
	flag.StringVar(&port, "p", "", "port or port range. 80. 1-65535 | 21,22,25 | 8080")
	flag.StringVar(&file, "f", "", "load file path")
	flag.IntVar(&timeout, "timeout", 2, "connection timeout")
	flag.IntVar(&threads, "t", 200, "number of concurrent threads")
	flag.StringVar(&outputFile, "o", "", "result output file path")
	flag.StringVar(&method, "method", "GET", "request method. -method GET | POST ...")
	flag.StringVar(&requestBody, "body", "", "post body. -body a=1&b=2")
	flag.StringVar(&bodyFile, "bodyfile", "", "post body file. -bodyfile ./post.txt")
	flag.StringVar(&path, "path", "/", "request url path. -path /phpinfo.php")
	flag.BoolVar(&redirect, "redirect", false, "follow 30x redirect")
	flag.Var(&reqHeaders, "H", "request headers. exmaple: -H User-Agent: xx -H Referer: xx")
	flag.StringVar(&grepString, "grep", "", "response body grep string. -grep phpinfo")
	flag.StringVar(&filterString, "filter", "", "response grep string. -filter Apache")
	flag.IntVar(&code, "code", 0, "response status code grep. -code 200")
	flag.StringVar(&proxies, "x", "", "set request proxy. -x socks://127.0.0.1:1080 | http://127.0.0.1:1086")
	flag.BoolVar(&isReplace, "replace", false, "use {{scheme}} {{host}} {{hostname}} {{path}} template string")
	flag.BoolVar(&debug, "debug", false, "print debug info")
	flag.BoolVar(&forceSSL, "forcessl", false, "force usage of SSL/HTTPS")
	flag.BoolVar(&displayProgress, "pg", false, "display progress bar")
	flag.BoolVar(&displayResponseBdoy, "response", false, "display response body")
	flag.BoolVar(&random, "random", false, "random request")
	flag.StringVar(&extractInfoReStr, "regexp", "", "regular expression for extracting information")
	flag.Parse()

	if (host == "" || port == "") && file == "" {
		flag.Usage()
		os.Exit(0)
	}
}

func makeClient() *http.Client {
	// 不校验证书
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,
			Deadline:  time.Now().Add(time.Duration(timeout) * time.Second),
			KeepAlive: time.Duration(timeout) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}}

	// 配置代理
	if proxies != "" {
		proxyURL := getProxyURL(proxies)
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}

	if !redirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func makeHeaders() {
	// 处理请求头
	for _, line := range reqHeaders {
		pair := strings.SplitN(line, ":", 2)
		if len(pair) == 2 {
			k, v := pair[0], strings.Trim(pair[1], " ")
			if strings.ToLower(k) == "host" {
				reqHost = v
			}
			headers[k] = v
		}
	}
}

func main() {

	// 检查是否合法请求方法
	if !validMethod(strings.ToUpper(method)) {
		fmt.Printf("net/http: invalid method %q", method)
		os.Exit(0)
	}

	ch = make(chan bool, threads)
	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)
	addrList := []string{}

	if len(ipList) != 0 && len(portList) != 0 {
		for _, host := range ipList {
			for _, port := range portList {
				address := fmt.Sprintf("%s:%d", host, port)
				addrList = append(addrList, address)
			}
		}
	} else if file != "" {
		lines, err := common.ReadFileLines(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		lines = common.ParseLines(lines)

		for _, line := range lines {

			line = strings.TrimSpace(line)
			host := line
			port := 80

			if strings.Contains(line, ":") {
				hostPort := strings.Split(line, ":")
				host = hostPort[0]
				port, _ = strconv.Atoi(hostPort[1])
			}

			if len(portList) != 0 {
				for _, p := range portList {
					address := fmt.Sprintf("%s:%d", host, p)
					addrList = append(addrList, address)
				}
			} else {
				address := fmt.Sprintf("%s:%d", host, port)
				addrList = append(addrList, address)
			}
		}
	}

	if bodyFile != "" {
		dat, err := ioutil.ReadFile(bodyFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		requestBody = string(dat)
	}

	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		common.CheckError(err)
		defer f.Close()
	}

	if extractInfoReStr != "" {
		extractInfoRe = regexp.MustCompile(fmt.Sprintf(`(?is)%s`, extractInfoReStr))
	}

	// 打印所有参数
	common.PrintInfo(len(addrList))

	// 进度条
	if displayProgress {
		bar = pb.New(len(addrList))
		bar.ShowSpeed = false
		bar.ShowTimeLeft = false
		bar.Start()
	}

	if random {
		common.Shuffle(addrList)
	}

	makeHeaders()

	startTime := time.Now()
	for _, line := range addrList {
		ch <- true
		wg.Add(1)

		pair := strings.SplitN(line, ":", 2)
		host := pair[0]
		port, _ := strconv.Atoi(pair[1])

		var url string
		if !forceSSL {
			url = fmt.Sprintf("http://%s%s", line, path)
		} else {
			url = fmt.Sprintf("https://%s%s", line, path)
		}
		if port == 443 {
			url = fmt.Sprintf("https://%s%s", host, path)
		}
		if isValidURL(url) {
			go FetchResponse(url)
		} else {
			log.Printf("%s is not a valid url", url)
		}
	}
	wg.Wait()

	fmt.Println()
	fmt.Println(" Complete requests:  ", successCount+errorCount)
	fmt.Println(" Successful requests:", successCount)
	fmt.Println(" Failed requests:    ", errorCount)
	fmt.Println(" Match requests:     ", matchCount)
	fmt.Println()
	finishMessage := fmt.Sprintf(" Time taken for tests: %v\n\n", time.Since(startTime))
	if displayProgress {
		bar.FinishPrint(fmt.Sprintf(finishMessage))
	} else {
		fmt.Printf(finishMessage)
	}
}

func Base64Encode(s string) string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(s))
	return sEnc
}

func getVarMap(requestURL string) map[string]string {
	var varMap map[string]string
	u, _ := url.Parse(requestURL)
	varMap = map[string]string{
		"{{scheme}}":      u.Scheme,
		"{{host}}":        u.Host,
		"{{hostname}}":    u.Hostname(),
		"{{path}}":        u.Path,
		"{{base64_host}}": Base64Encode(host),
		"{{url}}":         requestURL,
		"{{base64_url}}":  Base64Encode(requestURL),
	}
	return varMap
}

func CheckError(err error) {
	atomic.AddUint64(&errorCount, 1)
	if strings.Contains(err.Error(), "too many open files") {
		color.Red.Printf("[err] %s\n", err)
		return
	}
	if debug {
		color.Red.Printf("[err] %s\n", err)
	}
}

func FetchResponse(url string) {
	defer func() {
		<-ch
		if bar != nil {
			bar.Increment()
		}
		wg.Done()
	}()

	var req *http.Request
	var err error

	nRequestBody := requestBody
	if isReplace {
		varMap := getVarMap(url)
		for k, v := range varMap {
			url = strings.ReplaceAll(url, k, v)
			nRequestBody = strings.ReplaceAll(nRequestBody, k, v)
		}
	}

	if method == http.MethodPost || (method == "GET" && nRequestBody != "") {
		req, err = http.NewRequest(http.MethodPost, url, strings.NewReader(nRequestBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	} else if method == http.MethodPut {
		req, err = http.NewRequest(method, url, strings.NewReader(nRequestBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil || req == nil {
		fmt.Println(err)
		return
	}

	if reqHost != "" {
		req.Host = reqHost
	}
	req.Close = true

	for k, v := range headers {
		req.Header.Set(k, v)

	}

	client := makeClient()
	resp, err := client.Do(req)
	if err != nil {
		CheckError(err)
		return
	}
	defer resp.Body.Close()

	httpInfo := HttpInfo{
		URL:        url,
		StatusCode: resp.StatusCode,
	}

	// 获取编码
	reader := bufio.NewReader(resp.Body)
	e := determineEncoding(reader)
	utf8Reader := transform.NewReader(reader, e.NewDecoder())

	// 获取标题
	body, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		CheckError(err)
		body = []byte("")
	}
	respBody := string(body)
	atomic.AddUint64(&successCount, 1)

	// 提取标题
	m := titleRe.FindStringSubmatch(respBody)
	if len(m) >= 2 {
		httpInfo.Title = strings.TrimSpace(m[1])
	}

	if extractInfoRe != nil {
		// 正则提取额外信息
		m2 := extractInfoRe.FindStringSubmatch(respBody)
		if len(m2) >= 2 {
			httpInfo.ExtraInfo = strings.TrimSpace(m2[1])
		}
	}

	// 从响应头中提取字段 Server Content-Type X-Powered-By
	httpInfo.Server = resp.Header.Get("Server")
	httpInfo.ContentLength = resp.Header.Get("Content-Length")
	httpInfo.XPoweredBy = resp.Header.Get("X-Powered-By")
	pair := strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)
	if len(pair) == 2 {
		httpInfo.ContentType = pair[0]
	}
	result = append(result, httpInfo)

	statusCode := strconv.Itoa(httpInfo.StatusCode)

	// 通过响应头信息筛选响应 (response body. server httpInfo. status code)
	if strings.Contains(respBody, grepString) && strings.Contains(httpInfo.String(), filterString) && (code == 0 || strings.HasPrefix(statusCode, strconv.Itoa(code))) {
		var line = fmt.Sprintf("%-5d %-6s %-16s %-68s %-21s %-50s %s\n", httpInfo.StatusCode, httpInfo.ContentLength, httpInfo.ContentType, httpInfo.Server, httpInfo.XPoweredBy, httpInfo.URL, httpInfo.Title)
		writeLine := line
		atomic.AddUint64(&matchCount, 1)

		if strings.HasPrefix(statusCode, "2") {
			line = color.Green.Sprint(line)
		} else if strings.HasPrefix(statusCode, "3") {
			line = color.Magenta.Sprint(line)
		} else if strings.HasPrefix(statusCode, "4") {
			line = color.Yellow.Sprint(line)
		} else if strings.HasPrefix(statusCode, "5") {
			line = color.Red.Sprint(line)
		}

		if httpInfo.ExtraInfo != "" {
			extractInfoLine := fmt.Sprintf("%s\n", httpInfo.ExtraInfo)
			line += color.LightBlue.Sprint(extractInfoLine)
			writeLine += extractInfoLine
		}

		if displayResponseBdoy {
			respBodyLine := fmt.Sprintf("%s\n", respBody)
			line += color.LightBlue.Sprint(respBodyLine)
			writeLine += respBodyLine
		}

		fmt.Printf(line)
		f.WriteString(writeLine)
	}
}

func isValidURL(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}
