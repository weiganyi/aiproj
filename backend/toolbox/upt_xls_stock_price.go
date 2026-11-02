package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/xuri/excelize/v2"
)

// 全局共享的 http client（带 Cookie Jar），避免每次都重新获取 Cookie
var (
	httpClient *http.Client
	// xqJar 与 httpClient 配套的 CookieJar，用于注入 WAF 挑战算出的 cookie
	xqJar *cookiejar.Jar
	// plainHTTPClient 不带 CookieJar，用于调用方显式传入 cookie 的场景，
	// 避免 jar 中的 cookie 与手动设置的 Cookie 头重复叠加
	plainHTTPClient *http.Client
	userAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

// 阿里云 WAF（anti-bot）acw_sc__v2 JS 挑战的参数：
// 第一次请求 xueqiu.com 时，服务器可能不下发 xq_a_token，而是返回一段混淆 JS，
// 其中含 var arg1='xxxxxxxx'；必须按 unsbox+hexXor 算出 acw_sc__v2 cookie，
// 带着 acw_tc + acw_sc__v2 再请求一次，才能拿到 xq_a_token 等 cookie。
var acwPosList = []int{
	0xF, 0x23, 0x1D, 0x18, 0x21, 0x10, 0x1, 0x26, 0xA, 0x9,
	0x13, 0x1F, 0x28, 0x1B, 0x16, 0x17, 0x19, 0xD, 0x6, 0xB,
	0x27, 0x12, 0x14, 0x8, 0xE, 0x15, 0x20, 0x1A, 0x2, 0x1E,
	0x7, 0x4, 0x11, 0x5, 0x3, 0x1C, 0x22, 0x25, 0xC, 0x24,
}
const acwMask = "3000176000856006061501533003690027800375"

// 从 WAF 挑战页面提取 arg1（40 位十六进制）
var acwArg1Re = regexp.MustCompile(`var arg1='([0-9A-Fa-f]{40})'`)

// calcAcwScV2 计算 acw_sc__v2 cookie 的值（还原自阿里系混淆 JS）
func calcAcwScV2(arg1 string) string {
	// unsbox：对输入串 arg1 的每个字符，找到 posList 中值等于 i+1 的下标 j，
	// 将该字符放到输出串的第 j 位（posList 是 1-based 位置表）
	unpacked := make([]byte, len(arg1))
	for i := 0; i < len(arg1); i++ {
		for j, p := range acwPosList {
			if p == i+1 && j < len(unpacked) {
				unpacked[j] = arg1[i]
			}
		}
	}
	// hexXor：每两个 hex 字符作为一个字节，与 mask 对应字节异或，结果输出两位十六进制
	var sb strings.Builder
	for k := 0; k+1 < len(unpacked) && k+1 < len(acwMask); k += 2 {
		a, errA := strconv.ParseUint(string(unpacked[k:k+2]), 16, 32)
		b, errB := strconv.ParseUint(acwMask[k:k+2], 16, 32)
		if errA != nil || errB != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("%02x", a^b))
	}
	return sb.String()
}

// getPlainHTTPClient 返回一个不带 CookieJar 的共享 http client
func getPlainHTTPClient() *http.Client {
	if plainHTTPClient == nil {
		plainHTTPClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: &http.Transport{TLSHandshakeTimeout: 10 * time.Second},
		}
	}
	return plainHTTPClient
}

// jarHasToken 检查 jar 中是否已有 xq_a_token / xqat cookie
func jarHasToken() bool {
	if xqJar == nil {
		return false
	}
	u, _ := url.Parse("https://stock.xueqiu.com/")
	cookies := xqJar.Cookies(u)
	for _, c := range cookies {
		if c.Name == "xq_a_token" || c.Name == "xqat" {
			return true
		}
	}
	return false
}

// fetchWithHTTPClient 用指定 client GET 一个 URL，返回 body 内容和 HTTP 状态码（已读完并关闭 body）
func fetchWithHTTPClient(client *http.Client, u string) ([]byte, int, error) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("sec-ch-ua", `"Chromium";v="126", "Google Chrome";v="126"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// 注入一个 cookie 到 jar（domain 作用域为 .xueqiu.com，子域均可用）
func setJarCookie(name, value string) {
	u, _ := url.Parse("https://xueqiu.com/")
	xqJar.SetCookies(u, []*http.Cookie{{Name: name, Value: value, Domain: "xueqiu.com", Path: "/"}})
}

// 雪球风控挑战（腾讯 EdgeOne）: 被限流/风控时，服务器不下发 xq_a_token，
// 而是返回一段混淆 JS（有时包在完整 HTML 页面里），脚本会在浏览器中向
// document.cookie 写入 __tst_status=<求和数># 和 EO_Bot_Ssid=<数字> 并刷新页面。
// 这里用 goja 内嵌 JS 引擎执行该脚本、收集写入的 cookie 再注入 jar，重试即可正常换到 token。
var scriptRe = regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`)

// isEOChallenge 判断响应是否为 EO 风控挑战页
func isEOChallenge(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "__tst_status") || strings.Contains(s, "EO_Bot_Ssid")
}

// solveEOChallenge 用 goja 执行挑战 JS，返回它向 document.cookie 写入的赋值串列表
// （形如 "__tst_status=1585522647#"、"EO_Bot_Ssid=207814656"）
func solveEOChallenge(js string) ([]string, error) {
	vm := goja.New()
	// 防止异常脚本死循环：3 秒后中断（每次新建 VM，不影响后续调用）
	timer := time.AfterFunc(3*time.Second, func() {
		vm.Interrupt("挑战脚本执行超时")
	})
	defer timer.Stop()

	var assigned []string
	cur := ""
	doc := vm.NewObject()
	// 把 document.cookie 定义为 accessor，拦截每次赋值
	if err := doc.DefineAccessorProperty("cookie",
		vm.ToValue(func(fc goja.FunctionCall) goja.Value {
			return vm.ToValue(cur)
		}),
		vm.ToValue(func(fc goja.FunctionCall) goja.Value {
			s := fc.Argument(0).String()
			cur = s
			assigned = append(assigned, s)
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE); err != nil {
		return nil, err
	}
	vm.Set("document", doc)
	// setTimeout 的回调文案是 location.href 自刷新，无需真正执行，直接忽略
	vm.Set("setTimeout", func(fc goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	if _, err := vm.RunString(js); err != nil && len(assigned) == 0 {
		return nil, fmt.Errorf("执行风控挑战脚本失败: %w", err)
	}
	if len(assigned) == 0 {
		return nil, fmt.Errorf("风控挑战脚本未写入任何 cookie")
	}
	return assigned, nil
}

// parseCookieAssign 从 "name=value;attr..." 形式的 document.cookie 赋值串解析出 name 和 value
func parseCookieAssign(s string) (string, string) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ";"); i >= 0 {
		s = s[:i]
	}
	j := strings.Index(s, "=")
	if j < 0 {
		return "", ""
	}
	return strings.TrimSpace(s[:j]), strings.TrimSpace(s[j+1:])
}

// doInitHTTPClient 执行一轮"访问首页拿 Cookie"的引导流程，
// 逐步请求多家雪球 URL，最多 6 步：
// 1. 响应正常且已下发 xq_a_token -> 成功
// 2. 响应是 EO 风控挑战 -> 用 goja 计算应写入的 cookie 注入 jar 后重试
// 3. 响应是阿里系 acw_sc__v2 挑战 -> 按固定算法计算注入 jar 后重试
func doInitHTTPClient() error {
	bootstrapURLs := []string{
		"https://xueqiu.com/",
		"https://www.xueqiu.com/",
		"https://xueqiu.com/hq",
	}
	for step := 0; step < 6; step++ {
		bu := bootstrapURLs[step%len(bootstrapURLs)]
		body, _, err := fetchWithHTTPClient(httpClient, bu)
		if err != nil {
			continue
		}
		if jarHasToken() {
			return nil
		}

		switch {
		case isEOChallenge(body):
			// 挑战可能包在完整 HTML 页面里，先剥出 script 内容
			js := string(body)
			if sm := scriptRe.FindStringSubmatch(string(body)); len(sm) > 1 {
				js = sm[1]
			}
			assigned, serr := solveEOChallenge(js)
			if serr != nil {
				fmt.Printf("风控挑战处理失败 (%s): %v\n", bu, serr)
				continue
			}
			var list []*http.Cookie
			for _, a := range assigned {
				name, value := parseCookieAssign(a)
				if name == "" {
					continue
				}
				list = append(list, &http.Cookie{Name: name, Value: value, Domain: "xueqiu.com", Path: "/"})
			}
			xu, _ := url.Parse("https://xueqiu.com/")
			xqJar.SetCookies(xu, list)
		case acwArg1Re.Match(body):
			// 阿里系 acw_sc__v2 变体：从页面提取 arg1，按固定算法算出值注入 jar
			m := acwArg1Re.FindSubmatch(body)
			v2 := calcAcwScV2(string(m[1]))
			setJarCookie("acw_sc__v2", v2)
		default:
			// 其他响应：既没有 token 也没有挑战，继续请求下一个 URL
		}
	}
	if jarHasToken() {
		return nil
	}
	return fmt.Errorf("获取 Cookie 失败，未拿到 xq_a_token，请稍后重试")
}

// initHTTPClient 初始化 http client，并访问雪球首页获取必要的 Cookie。
// force 为 true 时忽略已有 client，强制重新获取（用于 cookie 失效后的重建）
func initHTTPClient(force bool) error {
	if httpClient != nil && !force {
		return nil
	}
	jar, _ := cookiejar.New(nil)
	xqJar = jar
	httpClient = &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSHandshakeTimeout: 10 * time.Second},
	}
	// 共重试 3 轮，每轮之间稍作等待，尽量在限流/风控下也能成功
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		if err := doInitHTTPClient(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

// GetWebValFromKey 根据股票代码从雪球网站获取当前股价
// 入参: stockCode 股票代码 (如 "00700" 不带市场前缀)
//
//	cookie     调用方提供的 Cookie 字符串（形如 "xq_a_token=xxx; u=yyy"），
//	             为空时回退到自动访问雪球首页获取 Cookie 的旧逻辑
//
// 返回: 股价字符串
func GetWebValFromKey(sec int, stockCode, cookie string) (string, error) {
	cookie = strings.TrimSpace(cookie)

	// 选择使用的 http client：显式传入 cookie 时用不带 jar 的 client，
	// 否则沿用自动获取 Cookie 的共享 client
	var client *http.Client
	if cookie != "" {
		client = getPlainHTTPClient()
	} else {
		if err := initHTTPClient(false); err != nil {
			return "", err
		}
		client = httpClient
	}

	// 使用雪球实时行情 JSON API（不是 HTML 页面）
	targetURL := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/quote.json?symbol=%s&extend=detail", stockCode)

	// 最多尝试 2 次：第 1 次失败后可能是 Cookie 失效/被 WAF 拦截，
	// 重新走一遍 Cookie 引导流程再试
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 && cookie == "" {
			if err := initHTTPClient(true); err != nil {
				break
			}
		}

		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return "", fmt.Errorf("构造请求失败: %w", err)
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Referer", "https://xueqiu.com/S/"+stockCode)
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Origin", "https://xueqiu.com")
		req.Header.Set("sec-ch-ua", `"Chromium";v="126", "Google Chrome";v="126"`)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", `"Windows"`)
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		// 成功获取响应后延时 sec 秒，避免请求过于频繁被雪球限流
		time.Sleep(time.Second * time.Duration(sec))

		// 使用正则匹配 JSON 中的 "current" 字段值
		// API 返回格式类似：{"data":{"quote":{"current":448.0,...}},...}
		re := regexp.MustCompile(`"current"\s*:\s*([0-9.]+)`)
		match := re.FindSubmatch(body)
		if match == nil {
			lastErr = fmt.Errorf("%s 未找到 current 字段，HTTP %d，响应: %s", targetURL, resp.StatusCode, truncateString(string(body), 200))
			continue
		}
		return string(match[1]), nil
	}
	return "", lastErr
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// UptXlsStockPrice 读取指定的 xlsx 文件，找到"代码"这一列，
// 依次调用 GetWebValFromKey 获取股价，并把股价更新到对应行的"股价"列单元格
// cookie 为可选参数，传入后会原样带给雪球接口；不传则自动获取 Cookie
func UptXlsStockPrice(sec int, xlsxFile string, cookie ...string) error {
	ck := ""
	if len(cookie) > 0 {
		ck = cookie[0]
	}

	f, err := excelize.OpenFile(xlsxFile)
	if err != nil {
		return fmt.Errorf("打开文件 %s 失败: %w", xlsxFile, err)
	}
	defer f.Close()

	// 取第一个 Sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return fmt.Errorf("文件中未找到工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取行失败: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("工作表为空")
	}

	// 在表头中查找"代码"列和"股价"列和"持仓"列
	codeColIdx := -1
	priceColIdx := -1
	positionColIdx := -1
	for i, cell := range rows[0] {
		switch cell {
		case "代码":
			codeColIdx = i
		case "股价":
			priceColIdx = i
		case "持仓":
			positionColIdx = i
		}
	}
	if codeColIdx == -1 {
		return fmt.Errorf("未在表头中找到\"代码\"列")
	}
	if priceColIdx == -1 {
		return fmt.Errorf("未在表头中找到\"股价\"列")
	}

	// 从第二行开始遍历，取出"代码"列的值
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if codeColIdx >= len(row) || positionColIdx >= len(row) {
			continue
		}
		code := row[codeColIdx]
		if code == "" {
			continue
		}
		// 跳过持仓为0的行
		if positionColIdx != -1 {
			positionStr := row[positionColIdx]
			position, err := strconv.Atoi(positionStr)
			if err == nil && position == 0 {
				continue
			}
		}

		price, err := GetWebValFromKey(sec, code, ck)
		if err != nil {
			fmt.Printf("code:%s, ck:%s, 获取失败 (%v)\n", code, ck, err)
			continue
		}
		fmt.Printf("%s：%s\n", code, price)

		// 将股价写入对应行的"股价"列
		// excelize 的行列号从 1 开始，i 是 0-based 的行索引，所以单元格行号为 i+1
		cellName, err := excelize.CoordinatesToCellName(priceColIdx+1, i+1)
		if err != nil {
			fmt.Printf("%s：生成单元格坐标失败 (%v)\n", code, err)
			continue
		}
		if err := f.SetCellValue(sheetName, cellName, price); err != nil {
			fmt.Printf("%s：写入单元格 %s 失败 (%v)\n", code, cellName, err)
			continue
		}
	}

	// 先在 xls 中定位"更新日期"单元格所在的行和列，
	// 然后在该行中、列号大于"更新日期"的单元格里查找 YYYYMMDD 样式的单元格，
	// 并将其值更新为今天的日期
	today := time.Now().Format("20060102")
	dateRe := regexp.MustCompile(`^\d{8}$`)
	dateUpdated := false
	for i, row := range rows {
		// 在当前行查找"更新日期"标签所在列
		labelColIdx := -1
		for j, cell := range row {
			if cell == "更新日期" {
				labelColIdx = j
				break
			}
		}
		if labelColIdx == -1 {
			continue
		}
		// 仅在"更新日期"之后的列中查找 YYYYMMDD 样式的单元格
		for j := labelColIdx + 1; j < len(row); j++ {
			cell := row[j]
			if !dateRe.MatchString(cell) {
				continue
			}
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				fmt.Printf("生成更新日期单元格坐标失败 (%v)\n", err)
				continue
			}
			if err := f.SetCellValue(sheetName, cellName, today); err != nil {
				fmt.Printf("写入更新日期单元格 %s 失败 (%v)\n", cellName, err)
				continue
			}
			fmt.Printf("更新日期已写入 %s: %s -> %s\n", cellName, cell, today)
			dateUpdated = true
			break
		}
		if dateUpdated {
			break
		}
	}
	if !dateUpdated {
		fmt.Println("未找到\"更新日期\"后存在 YYYYMMDD 样式单元格的行，跳过日期更新")
	}

	// 更新带有公式的单元格的值（刷新计算缓存）
	if err := f.UpdateLinkedValue(); err != nil {
		fmt.Printf("更新公式单元格的缓存值失败: %v\n", err)
	}

	// 保存文件
	if err := f.Save(); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}
