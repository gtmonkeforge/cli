package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"
)

type Session struct {
	Client      *http.Client
	CSRFToken   string
	User        json.RawMessage
	credentials *credentialJar
}

func Login(ctx context.Context, openBrowser func(string) error) (*Session, error) {
	saved, found, err := loadSession(ctx)
	if err != nil || found {
		return saved, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	verifier, err := randomValue()
	if err != nil {
		return nil, err
	}
	state, err := randomValue()
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for login: %w", err)
	}
	defer listener.Close()
	redirectURI := "http://" + listener.Addr().String() + "/callback"
	codes := make(chan string, 1)
	var accepted atomic.Bool
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Expected GET", http.StatusMethodNotAllowed)
			return
		}
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || len(q["state"]) != 1 || len(q["code"]) != 1 {
			http.Error(w, "Invalid login callback", http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(state)) != 1 || !validCode(code) {
			http.Error(w, "Invalid login callback", http.StatusBadRequest)
			return
		}
		if !accepted.CompareAndSwap(false, true) {
			http.Error(w, "Login callback already received", http.StatusConflict)
			return
		}
		codes <- code
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Login callback received. Return to the CLI to finish signing in.")
	})
	server := &http.Server{
		Handler: mux, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	defer server.Close()
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()
	digest := sha256.Sum256([]byte(verifier))
	query := url.Values{
		"redirect_uri": {redirectURI}, "state": {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(digest[:])},
		"code_challenge_method": {"S256"},
	}
	if openBrowser == nil {
		openBrowser = OpenBrowser
	}
	if err := openBrowser(Origin + "/auth/discord?" + query.Encode()); err != nil {
		return nil, fmt.Errorf("open login browser: %w", err)
	}
	var code string
	select {
	case code = <-codes:
	case err := <-serveErrors:
		return nil, fmt.Errorf("login callback server: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for login: %w", ctx.Err())
	}
	jar, err := newCredentialJar()
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar, Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	body, err := json.Marshal(map[string]string{
		"code": code, "code_verifier": verifier, "redirect_uri": redirectURI,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		Origin+"/api/v1/auth/native/exchange", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange login code: %w", err)
	}
	err = decodeSession(resp, nil)
	if err != nil {
		if errors.Is(err, ErrLoggedOut) {
			return nil, clearInvalidCredentials()
		}
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, Origin+"/api/v1/session", nil)
	if err != nil {
		return nil, err
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verify login session: %w", err)
	}
	result := &Session{Client: client, credentials: jar}
	if err := decodeSession(resp, result); err != nil {
		if errors.Is(err, ErrLoggedOut) {
			return nil, clearInvalidCredentials()
		}
		return nil, err
	}
	if err := result.CredentialError(); err != nil {
		return nil, err
	}
	return result, nil
}

func decodeSession(resp *http.Response, result *Session) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return ErrLoggedOut
		}
		return fmt.Errorf("MonkeForge authentication returned HTTP %d", resp.StatusCode)
	}
	var data struct {
		User      json.RawMessage `json:"user"`
		CSRFToken string          `json:"csrf_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&data); err != nil {
		return fmt.Errorf("decode login session: %w", err)
	}
	if len(data.User) == 0 || bytes.Equal(bytes.TrimSpace(data.User), []byte("null")) || data.CSRFToken == "" {
		return ErrLoggedOut
	}
	if result != nil {
		result.User, result.CSRFToken = data.User, data.CSRFToken
	}
	return nil
}

func randomValue() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate login randomness: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func validCode(code string) bool {
	if len(code) != 43 {
		return false
	}
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func OpenBrowser(loginURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", loginURL)
	case "linux":
		cmd = exec.Command("xdg-open", loginURL)
	default:
		return fmt.Errorf("unsupported browser opener on %s; supply an openBrowser callback", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
