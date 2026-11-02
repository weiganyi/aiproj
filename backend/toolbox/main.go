package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: toolbox <功能> [参数...]")
		fmt.Println("   <xls文件路径> - 调用 UptXlsStockPrice 更新指定 xlsx 文件的股价")
		return
	}

	switch os.Args[1] {
	case "uxls":
		// 兼容两种格式: uxls sec <xls文件路径> [cookie]  和旧格式 uxls <xls文件路径> [cookie]（默认 sec=3）
		sec := 3
		idx := 2
		if len(os.Args) >= 3 {
			if s, err := strconv.Atoi(os.Args[2]); err == nil {
				sec = s
				idx = 3
			}
		} else {
			fmt.Println("入参错误：请至少提供 xlsx 文件路径")
			fmt.Println("用法: toolbox uxls sec <xls文件路径> [cookie]")
			return
		}
		if len(os.Args) <= idx {
			fmt.Println("入参错误：请提供 xlsx 文件路径")
			fmt.Println("用法: toolbox uxls sec <xls文件路径> [cookie]")
			return
		}
		xlsxFile := os.Args[idx]
		cookie := ""
		if len(os.Args) > idx+1 {
			cookie = os.Args[idx+1]
		}
		if err := UptXlsStockPrice(sec, xlsxFile, cookie); err != nil {
			fmt.Println("UptXlsStockPrice 执行失败:", err)
		}
	default:
		fmt.Println("未知的功能:", os.Args[1])
	}
}
