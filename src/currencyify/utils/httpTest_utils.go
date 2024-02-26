package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (r *ExternalRequest) DoMock() (*http.Response, error) {
	apiNameSuffix := fmt.Sprintf("-%v", strings.ToLower(r.Name))
	mockType, ok := r.Headers[fmt.Sprintf("x-mock-api%v", apiNameSuffix)]
	if !ok {
		apiNameSuffix = ""
		mockType = r.Headers["x-mock-api"]
	}

	var rr = httptest.NewRecorder()

	var err error

	switch mockType {
	case "error_response":
		rr.WriteHeader(400)
		_, _ = rr.WriteString(`{"errors":"some error"}`)
	default:
		switch r.Name {
		case "GetLatestCurrencyRate":
			rr.WriteHeader(200)
			_, _ = rr.WriteString(`{"success":true,"terms":"https://fxratesapi.com/legal/terms-conditions","privacy":"https://fxratesapi.com/legal/privacy-policy","timestamp":1708949040,"date":"2024-02-26T12:04:00.000Z","base":"USD","rates":{"INR":82.771291,"JPY":150.608807,"USD":1}}`)
		case "GetLatestCurrencyExchangeRates":
			rr.WriteHeader(200)
			_, _ = rr.WriteString(`{"success":true,"terms":"https://fxratesapi.com/legal/terms-conditions","privacy":"https://fxratesapi.com/legal/privacy-policy","timestamp":1708949040,"date":"2024-02-26T12:04:00.000Z","base":"USD","rates":{"INR":82.771291,"JPY":150.608807}}`)
		default:
			err = errors.New("No matching API found")
		}
	}

	return rr.Result(), err
}

// GetMockHeadersFromContext fetches mock headers from request context and add it to the request headers.
func (r *ExternalRequest) GetMockHeadersFromContext() {
	if r.ReqCtx == nil {
		return
	}

	mockHeaders, ok := r.ReqCtx.Value(fmt.Sprintf("x-mock-headers-%v", strings.ToLower(r.Name))).(map[string]string)
	if !ok {
		mockHeaders, _ = r.ReqCtx.Value("x-mock-headers").(map[string]string)
	}
	if len(mockHeaders) <= 0 {
		return
	}

	for k, v := range mockHeaders {
		r.Headers[k] = v
	}
}
