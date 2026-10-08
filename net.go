package main

import (
	"fmt"
	"io"
	"net/http"
)

type ProgressFunc func(downloaded, total int64)

func DownloadString(url string) (string, error) {
	stream, err := DownloadStream(url)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func DownloadStream(url string) (io.ReadCloser, error) {
	return DownloadStreamWithProgress(url, nil)
}

// DownloadStreamWithProgress reports bytes read.
// total is -1 when the response length is unknown.
// The caller must close the returned stream.
func DownloadStreamWithProgress(
	url string,
	onProgress ProgressFunc,
) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// we handle 404s manually
	if (resp.StatusCode < 200 || resp.StatusCode >= 300) && resp.StatusCode != 404 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	if onProgress != nil {
		onProgress(0, resp.ContentLength)
	}

	return &progressStream{
		ReadCloser: resp.Body,
		total:      resp.ContentLength,
		onProgress: onProgress,
	}, nil
}

type progressStream struct {
	io.ReadCloser
	downloaded int64
	total      int64
	onProgress ProgressFunc
}

func (s *progressStream) Read(p []byte) (int, error) {
	n, err := s.ReadCloser.Read(p)
	s.downloaded += int64(n)

	if n > 0 && s.onProgress != nil {
		s.onProgress(s.downloaded, s.total)
	}

	return n, err
}
