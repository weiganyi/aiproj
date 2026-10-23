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
		if len(os.Args) < 4 {
			fmt.Println("用法: toolbox uxls sec <xls文件路径> [cookie]")
			return
		}
		sec, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("入参错误，sec:", sec)
		}
		xlsxFile := os.Args[3]
		cookie := ""
		if len(os.Args) > 4 {
			cookie = os.Args[4]
		}
		if err := UptXlsStockPrice(sec, xlsxFile, cookie); err != nil {
			fmt.Println("UptXlsStockPrice 执行失败:", err)
		}
	default:
		fmt.Println("未知的功能:", os.Args[1])
	}
}
