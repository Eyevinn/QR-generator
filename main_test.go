package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQrSize(t *testing.T) {
	s := &server{
		port: "8080",
		text: "sample",
	}

	handler := s.makeGenerateQRCodeHandler()

	tests := []struct {
		TestName           string
		QrSize             string
		ExpectedStatusCode int
	}{
		{
			TestName:           "Should return 400, for non integer qr size",
			QrSize:             "abc",
			ExpectedStatusCode: 400,
		},
		{
			TestName:           "Should return 400, for qr size less than 128",
			QrSize:             "100",
			ExpectedStatusCode: 400,
		},
		{
			TestName:           "Should return 400, for qr size greater than 2048",
			QrSize:             "3000",
			ExpectedStatusCode: 400,
		},
		{
			TestName:           "Should return 200, for valid qr size",
			QrSize:             "300",
			ExpectedStatusCode: 200,
		},
	}

	for _, test := range tests {
		url := "/?text=happy-coding&size=" + test.QrSize
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != test.ExpectedStatusCode {
			t.Fatalf("Expected status code %d, got %d", test.ExpectedStatusCode, rec.Code)
		}
	}
}
