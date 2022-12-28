package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type GerritService struct {
	url      string
	username string
	password string
}

const gerritResponsePrefix = ")]}'\n"

// NewGerrit creates a Gerrit service
func NewGerrit(url, username, password string) *GerritService {
	return &GerritService{
		url:      url,
		username: username,
		password: password,
	}
}

// ListFilesInChange lists changed files in a change.
// Docs: https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#list-files
func (s *GerritService) ListFilesInChange(ctx context.Context, changeKey, revisionKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/a/changes/%s/revisions/%s/files", s.url, changeKey, revisionKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	resp, err := parseGerritResponse(bytes)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// GetFileContent returns the file content in a change.
// Docs: https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-content
func (s *GerritService) GetFileContent(ctx context.Context, changeKey, revisionKey, filename string) (string, error) {
	url := fmt.Sprintf("%s/a/changes/%s/revisions/%s/files/%s/content", s.url, changeKey, revisionKey, url.QueryEscape(filename))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	bytes, err := s.doRequest(req)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func (s *GerritService) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", s.basicAuth()))

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

func (s *GerritService) basicAuth() string {
	auth := fmt.Sprintf("%s:%s", s.username, s.password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// All responses from Gerrit have a specific prefix
// Docs: https://gerrit-review.googlesource.com/Documentation/rest-api.html
func parseGerritResponse(input []byte) ([]byte, error) {
	str := string(input)
	if !strings.HasPrefix(str, gerritResponsePrefix) {
		return nil, errors.Errorf("invalid response, missing prefix %s", gerritResponsePrefix)
	}

	sec := strings.Split(str, gerritResponsePrefix)
	if len(sec) != 2 {
		return nil, errors.Errorf("invalid response %s", str)
	}

	return []byte(sec[1]), nil
}
