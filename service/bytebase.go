package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/pkg/errors"
)

type BytebaseService struct {
	url    string
	key    string
	secret string
}

type bytebaseAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type bytebaseAuthResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

// NewLark creates a Bytebase service
func NewBytebase(url, key, secret string) *BytebaseService {
	return &BytebaseService{
		url:    url,
		key:    key,
		secret: secret,
	}
}

func (s *BytebaseService) CreateIssue(ctx context.Context, create *payload.IssueCreate) error {
	payload, err := json.Marshal(create)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/issue", s.url), strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	if _, err := s.doRequest(req); err != nil {
		return err
	}

	return nil
}

func (s *BytebaseService) login() (*bytebaseAuthResponse, error) {
	rb, err := json.Marshal(&bytebaseAuthRequest{
		Email:    s.key,
		Password: s.secret,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/auth/login", s.url), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := s.doRequestWithToken(req, "")
	if err != nil {
		return nil, err
	}

	res := &bytebaseAuthResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *BytebaseService) doRequest(req *http.Request) ([]byte, error) {
	user, err := s.login()
	if err != nil {
		return nil, err
	}

	return s.doRequestWithToken(req, user.Token)
}

func (s *BytebaseService) doRequestWithToken(req *http.Request, token string) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
