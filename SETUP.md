#  Hướng dẫn Triển khai Hệ thống SENT SOC
Bước 1: muốn triển khai yêu cầu máy phải có git và cần phải cài docker.
Bước 2: pull dự án bên gỉthub về với 1 trong 3 lệnh sau khuyến khích dùng SSH hoặc HTTPS lệnh sau
SSH : git@github.com:Gamlio/SENT-System.git
HTTPS : https://github.com/Gamlio/SENT-System.git
Github CLI:  gh repo clone Gamlio/SENT-System

Bước 3:  tạo .env với cấu trúc sau

# DATABASE_URL=postgresql:
# có thể tuỳ biến theo dự án và bảo mật
DATABASE_URL=
# cổng ở đây đặt không đáng kể vì chúng ta sẽ chạy ở docker
MONGODB_URI=
 # host mặc định
REDIS_HOST=redis
 # port mặc định
REDIS_PORT=6379
 # có thể đặt hoặc không
REDIS_PASSWORD=
REDIS_DB=0
#  --- SMTP CONFIGURATION ---
# địa chỉ email
SMTP_EMAIL= 
 # đặt pass theo email cung cấp
SMTP_PASSWORD=
# 2 cái này giữ nguyên
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587

ALGORITHM=HS256
# Domain mặc định
APP_URL= 
VITE_APP_WS_URL=wss://DomainName:8000/api/v1/ws
# thời gian truy cập nên đổi theo sự phù hợp
ACCESS_TOKEN_EXPIRE_MINUTES=1440
# bạn thích đặt gì thì đặt
JWT_SECRET= 
# bạn thích đặt gì thì đặt
INTERNAL_UPDATE_TOKEN=
# đây là nơi cho phép nên hạn chế mở linh tinh
ALLOWED_ORIGINS= 

PORT=8000
# có thể thêm môi trường là gì
ENVIRONMENT=
 # copy mã bên Tunnel để deploy
CF_TUNNEL_TOKEN=

Bước 4: mở terminal và chạy lệnh sau:
## nếu máy yếu chạy theo cách sau:
docker-compose build auth-service
docker-compose build user-service
docker-compose build asset-service 
docker-compose build ai-service
docker-compose build document-service
docker-compose build incident-service
docker-compose build policy-service
docker-compose build behavior-service
docker-compose build scoring-service
docker-compose build approval-service
docker-compose build groups-service
docker-compose build dashboard-service
docker-compose build frontend 
docker-compose up -d
# Nếu máy khoẻ có thể chạy lệnh sau:
docker-compose up -d --build

## lưu ý bên Linux hoặc các bản phân phối khác nên thên tiền tố sudo vào trước hoặc cấu hình riêng cho không cần sudo với docker (Không khuyến khích)


## Vậy là website đã chạy.
## Còn phần mềm hướng dẫn sử dụng các tải sẽ cập nhật sau.
cd SENT

# --- Dành cho Linux / macOS (Bash / Zsh) ---
# Build cho Windows (.exe)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/SENT_v4.2.6_windows.exe ./cmd

# Build cho Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/SENT_v4.2.6_linux ./cmd

# Build cho MacOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/SENT_v4.2.6_mac ./cmd

# --- Dành cho Windows (PowerShell) ---
# Build cho Windows (.exe)
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o dist/SENT_v4.2.6_windows.exe ./cmd

# Build cho Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o dist/SENT_v4.2.6_linux ./cmd

# Build cho MacOS
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o dist/SENT_v4.2.6_mac ./cmd



# nếu mà muốn đẩy SENT từ windows qua linux qua ssh làm như sau:
## Mở file main.go trên máy Windows, di chuyển hàm forkLinuxDaemon() xuống phía dưới (ngay sau khối lệnh config.LoadOrBootstrap).

## Chạy file build.sh (hoặc chạy lệnh go build) trên Windows để sinh ra file thực thi mới:

Bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/SENT_v4.2.6_linux cmd/main.go

## Bash
scp "địa chỉ chứa phần mềm\SENT_v4.2.6_linux" thanh@192.168.x.x:~
## 3. Kích hoạt và Nhập mã trên Debian
## SSH vào máy Debian: 
ssh thanh@192.168.x.x.

 ## Di chuyển vào thư mục Home: 
 cd ~.

 ## Cấp lại quyền thực thi cho file mới nhận: 
 chmod +x SENT_v4.2.6_linux.

 ## Khởi chạy phần mềm bằng quyền root:

Bash
sudo ./SENT_v4.2.6_linux
## Lúc này màn hình 🛡️ KÍCH HOẠT SENT SENSOR (CLI MODE) sẽ dừng lại và hiển thị dòng chữ bên dưới, bạn tiến hành nhập mã Token của mình vào rồi nhấn Enter:

## mã này lấy bên assets bên sent
>> Nhập Mã Cài Đặt (Enrollment Token):