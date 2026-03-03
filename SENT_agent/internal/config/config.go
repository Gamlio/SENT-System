package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const CONFIG_FILE = "agent_config.json"

type Config struct {
	CompanyCode string `json:"company_code"`
}

var Current Config

func Load() bool {
	file, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return false
	}
	json.Unmarshal(file, &Current)
	return Current.CompanyCode != ""
}

func Save(code string) {
	Current.CompanyCode = strings.TrimSpace(code)
	data, _ := json.Marshal(Current)
	os.WriteFile(CONFIG_FILE, data, 0644)
}

func PromptForCompanyCode() {
	reader := bufio.NewReader(os.Stdin) // Khai báo reader
	fmt.Println("===========================================")
	fmt.Println("   KÍCH HOẠT SENT AGENT (CLI MODE)")
	fmt.Println("===========================================")

	for {
		fmt.Print("👉 Nhập Mã Công Ty (VD: SME-XXXX): ")
		// Sử dụng biến reader ở đây
		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

		if len(code) >= 5 {
			Save(code)
			fmt.Println("✅ Kích hoạt thành công!")
			time.Sleep(1 * time.Second)
			break
		}
		fmt.Println("❌ Mã không hợp lệ.")
	}
}
