package analyzer

import (
	"fmt"
	"os/exec"
)

// 2. Hàm hiển thị Pop-up cảnh báo giữa màn hình người dùng
func ShowWindowsAlert(title, message string) {
	// Tận dụng PowerShell để gọi MessageBox của Windows (Rất nhẹ và không cần thư viện ngoài)
	psScript := fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('%s', '%s', 'OK', 'Warning')`, message, title)

	// Chạy lệnh PowerShell ngầm (-WindowStyle Hidden)
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", psScript)

	// Dùng Start() thay vì Run() để Pop-up hiện lên mà không làm treo (block) luồng chạy của Agent
	err := cmd.Start()
	if err != nil {
		fmt.Printf("⚠️ Lỗi hiển thị cảnh báo: %v\n", err)
	}
}
