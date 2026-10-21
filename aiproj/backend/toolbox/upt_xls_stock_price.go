package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// 全局共享的 http client（带 Cookie Jar），避免每次都重新获取 Cookie
var (
	httpClient *http.Client
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// initHTTPClient 初始化 http client，并访问雪球首页获取必要的 Cookie
func initHTTPClient() error {
	if httpClient != nil {
		return nil
	}
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{
		Jar:       jar,
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSHandshakeTimeout: 10 * time.Second},
	}

	// 依次访问雪球首页和一个行情页，让服务器下发 xq_a_token / xqat 等 Cookie
	// 雪球的 Set-Cookie 是通过 HTTP 响应头直接返回的，CookieJar 会自动处理
	bootstrapURLs := []string{
		"https://xueqiu.com/",
		"https://xueqiu.com/hq",
	}
	for _, u := range bootstrapURLs {
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("获取 Cookie 失败 (%s): %w", u, err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	// 检查 cookiejar 中是否拿到了 xq_a_token
	xueqiuURL, _ := http.NewRequest("GET", "https://xueqiu.com/", nil)
	cookies := jar.Cookies(xueqiuURL.URL)
	hasToken := false
	for _, c := range cookies {
		if c.Name == "xq_a_token" || c.Name == "xqat" {
			hasToken = true
			break
		}
	}
	if !hasToken {
		return fmt.Errorf("获取 Cookie 失败，未拿到 xq_a_token，请稍后重试")
	}

	return nil
}

// GetWebValFromKey 根据股票代码从雪球网站获取当前股价
// 入参: 股票代码 (如 "00700" 不带市场前缀)
// 返回: 股价字符串
func GetWebValFromKey(stockCode string) (string, error) {
	if err := initHTTPClient(); err != nil {
		return "", err
	}

	// 使用雪球实时行情 JSON API（不是 HTML 页面）
	targetURL := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/quote.json?symbol=%s&extend=detail", stockCode)

	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://xueqiu.com/S/"+stockCode)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Origin", "https://xueqiu.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// fmt.Println("API 响应:", string(body))
	// 成功获取响应后延时 1 秒，避免请求过于频繁被雪球限流
	time.Sleep(1 * time.Second)

	// 使用正则匹配 JSON 中的 "current" 字段值
	// API 返回格式类似：{"data":{"quote":{"current":448.0,...}},...}
	re := regexp.MustCompile(`"current"\s*:\s*([0-9.]+)`)
	match := re.FindSubmatch(body)
	if match == nil {
		return "", fmt.Errorf("%s 未找到 current 字段，响应: %s", targetURL, truncateString(string(body), 200))
	}
	return string(match[1]), nil
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// UptXlsStockPrice 读取指定的 xlsx 文件，找到"代码"这一列，
// 依次调用 GetWebValFromKey 获取股价，并把股价更新到对应行的"股价"列单元格
func UptXlsStockPrice(xlsxFile string) error {
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

	// 在表头中查找"代码"列和"股价"列
	codeColIdx := -1
	priceColIdx := -1
	for i, cell := range rows[0] {
		switch cell {
		case "代码":
			codeColIdx = i
		case "股价":
			priceColIdx = i
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
		if codeColIdx >= len(row) {
			continue
		}
		code := row[codeColIdx]
		if code == "" {
			continue
		}

		price, err := GetWebValFromKey(code)
		if err != nil {
			fmt.Printf("%s：获取失败 (%v)\n", code, err)
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

var _ = strings.TrimSpace
