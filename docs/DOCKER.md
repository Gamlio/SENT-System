🛠️ Scenario B: Low-Resource Optimization
Mandatory for machines with 8GB RAM (e.g., RTX 2050 4GB VRAM).

1. System Preparation (Windows Host)
Virtual Memory (Page File): Increase to 20GB - 30GB on an SSD (Drive D: recommended if C: is full). This acts as "backup RAM" to prevent Docker crashes.

Ollama Host Mode:

Close Ollama from the system tray.

Set Environment Variable: OLLAMA_HOST=0.0.0.0.

Restart Ollama.

Docker Resources: Assign at least 4GB - 6GB of RAM in Docker Desktop Settings -> Resources.

2. Sequential Build Strategy
Building 14 services simultaneously will crash an 8GB RAM machine. Use this "Step-by-Step" approach:

Step 1: Clean Build Cache

PowerShell
docker builder prune -f
Step 2: Build Frontend (Resource Intensive)

PowerShell
docker-compose build --no-deps frontend
Step 3: Build Backend Services in Groups

PowerShell
# Group 1: Core Services
docker-compose build auth-service user-service asset-service

# Group 2: AI & Documents
docker-compose build ai-service document-service policy-service

# Group 3: Remaining Services
docker-compose build behavior-service incident-service scoring-service software-service approval-service groups-service dashboard-service
Step 4: Launch the System

PowerShell
docker-compose up -d
🔍 Verification & Logs
Check Status: docker ps (Ensure all 14 containers are "Up").

Monitor Logs: docker logs -f ai_service to check AI connectivity.

Access Dashboard: Open http://localhost in your browser.

⚠️ Troubleshooting
RPC/EOF Error: Usually caused by RAM exhaustion. Restart Docker Desktop and ensure Page File is active.

Internal Server Error 500 (Docker Engine): The engine crashed. Restart Docker and build services one by one.

AI Connectivity: Ensure ai-service points to http://host.docker.internal:11434 if Ollama is running on the host.