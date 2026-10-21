package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: toolbox <功能> [参数...]")
		fmt.Println("   <xls文件路径> - 调用 UptXlsStockPrice 更新指定 xlsx 文件的股价")
		return
	}

	switch os.Args[1] {
	case "uxls":
		if len(os.Args) < 3 {
			fmt.Println("用法: toolbox uxls <xls文件路径>")
			return
		}
		xlsxFile := os.Args[2]
		if err := UptXlsStockPrice(xlsxFile); err != nil {
			fmt.Println("UptXlsStockPrice 执行失败:", err)
		}
	default:
		fmt.Println("未知的功能:", os.Args[1])
	}
}
