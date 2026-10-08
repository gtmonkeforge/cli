package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/zalando/go-keyring"
)

// Good luck jackass; uncommented & unshameful

const Origin = "https://monkeforge.org"

const credentialService = "mforge"
const credentialAccount = Origin

var ErrLoggedOut = errors.New("MonkeForge session is logged out")

type savedCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type credentialJar struct {
	http.CookieJar
	mu     sync.Mutex
	err    error
	closed bool
}

func newCredentialJar() (*credentialJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &credentialJar{CookieJar: jar}, nil
}

func (j *credentialJar) Cookies(u *url.URL) []*http.Cookie {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return nil
	}
	return j.CookieJar.Cookies(u)
}

func (j *credentialJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return
	}
	j.CookieJar.SetCookies(u, cookies)
	if u.Scheme != "https" || u.Host != "monkeforge.org" || len(cookies) == 0 {
		return
	}
	origin, _ := url.Parse(Origin)
	saved := []savedCookie{}
	for _, cookie := range j.CookieJar.Cookies(origin) {
		saved = append(saved, savedCookie{Name: cookie.Name, Value: cookie.Value})
	}
	if len(saved) == 0 {
		j.err = deleteCredentials()
		return
	}
	data, err := json.Marshal(saved)
	if err == nil {
		err = keyring.Set(credentialService, credentialAccount, string(data))
	}
	if err != nil {
		j.err = fmt.Errorf("save MonkeForge credentials: %w", err)
	} else {
		j.err = nil
	}
}

func (s *Session) CredentialError() error {
	if s.credentials == nil {
		return nil
	}
	s.credentials.mu.Lock()
	defer s.credentials.mu.Unlock()
	return s.credentials.err
}

func deleteCredentials() error {
	err := keyring.Delete(credentialService, credentialAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete MonkeForge credentials: %w", err)
	}
	return nil
}

func LoadSession(ctx context.Context) (*Session, error) {
	session, found, err := loadSession(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrLoggedOut
	}
	return session, nil
}

func loadSession(ctx context.Context) (*Session, bool, error) {
	secret, err := keyring.Get(credentialService, credentialAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, fmt.Errorf("load MonkeForge credentials: %w", err)
	}
	var saved []savedCookie
	if err := json.Unmarshal([]byte(secret), &saved); err != nil || len(saved) == 0 {
		return nil, true, clearInvalidCredentials()
	}
	jar, err := newCredentialJar()
	if err != nil {
		return nil, true, err
	}
	origin, _ := url.Parse(Origin)
	var cookies []*http.Cookie
	for _, savedCookie := range saved {
		cookie := &http.Cookie{Name: savedCookie.Name, Value: savedCookie.Value, Path: "/", Secure: true, HttpOnly: true}
		if cookie.Valid() != nil || cookie.Name == "" || cookie.Value == "" {
			return nil, true, clearInvalidCredentials()
		}
		cookies = append(cookies, cookie)
	}
	// Restore without writing before the server validates the session.
	jar.CookieJar.SetCookies(origin, cookies)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, Origin+"/api/v1/session", nil)
	if err != nil {
		return nil, true, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("validate saved MonkeForge login: %w", err)
	}
	result := &Session{Client: client, credentials: jar}
	if err := decodeSession(resp, result); err != nil {
		if errors.Is(err, ErrLoggedOut) {
			return nil, true, clearInvalidCredentials()
		}
		return nil, true, err
	}
	if err := result.CredentialError(); err != nil {
		return nil, true, err
	}
	return result, true, nil
}

func clearInvalidCredentials() error {
	if err := deleteCredentials(); err != nil {
		return fmt.Errorf("%w; %v", ErrLoggedOut, err)
	}
	return ErrLoggedOut
}

func Logout(ctx context.Context, session *Session) error {
	var requestErr error
	if session != nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, Origin+"/api/v1/logout", nil)
		if err == nil {
			req.Header.Set("X-CSRF-Token", session.CSRFToken)
			resp, err := session.Client.Do(req)
			if err != nil {
				requestErr = err
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					requestErr = fmt.Errorf("server logout returned HTTP %d", resp.StatusCode)
				}
			}
		} else {
			requestErr = err
		}
		if session.credentials != nil {
			session.credentials.mu.Lock()
			session.credentials.closed = true
			session.credentials.CookieJar, _ = cookiejar.New(nil)
			session.credentials.mu.Unlock()
		}
		session.CSRFToken, session.User = "", nil
	}
	if err := deleteCredentials(); err != nil {
		return err
	}
	return requestErr
}
