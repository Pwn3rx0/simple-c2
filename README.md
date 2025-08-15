# SIMPLE C2

Lightweight Command & Control (C2) setup with a **Flask server** and **Windows & linux Go agent** for educational and penetration testing purposes.

> **Warning:** Use only on systems you own or have explicit permission to test.

---

## Features

- Register multiple clients with unique IDs  
- Send commands to specific or all clients  
- Receive and display command output in real-time  
- Automatic cleanup of disconnected clients  
- Windows client runs commands via `cmd.exe` with timeout  

---

## Requirements

- **Server:** Python 3.8+, Flask  
- **Client:** Windows, Go 1.18+  

---

## Setup

### Server

```bash
git clone https://github.com/Pwn3rx0/simple-c2.git
cd simple-c2
pip install Flask
python c2.py
```
Access the web interface: http://0.0.0.0:8080
Client (Windows)
```
go build -o client.exe windows.go
```
Set SERVER_BASE in main.go to your server URL.
Usage

   - Start the server.
   - Run the client on target machine(s).
   - Send commands via the web interface.
   - View live outputs; use exit or quit to stop the client.

Contributing

    Improving UI, security, or client features

    Submit pull requests responsibly

