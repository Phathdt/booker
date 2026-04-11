package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type responseWrapperWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

func (w *responseWrapperWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *responseWrapperWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.body.Write(b)
}

var paginationKeys = map[string]bool{"total": true, "page": true, "limit": true}

func wrapListResponse(body []byte) []byte {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}
	var arrayKey string
	for k, v := range obj {
		if paginationKeys[k] {
			continue
		}
		trimmed := bytes.TrimSpace(v)
		if len(trimmed) > 0 && trimmed[0] == '[' {
			if arrayKey != "" {
				return nil
			}
			arrayKey = k
		} else {
			return nil
		}
	}
	if arrayKey == "" {
		return nil
	}
	result := map[string]json.RawMessage{"data": obj[arrayKey]}
	for k := range paginationKeys {
		if v, ok := obj[k]; ok {
			result[k] = v
		}
	}
	out, err := json.Marshal(result)
	if err != nil {
		return nil
	}
	return out
}

// WrapResponse wraps successful JSON responses in {"data": ...}.
func WrapResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := &responseWrapperWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapper, r)

		if wrapper.statusCode >= 400 {
			w.WriteHeader(wrapper.statusCode)
			_, _ = w.Write(wrapper.body.Bytes())
			return
		}

		if wrapper.written && wrapper.body.Len() > 0 {
			if listOut := wrapListResponse(wrapper.body.Bytes()); listOut != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(wrapper.statusCode)
				_, _ = w.Write(listOut)
				return
			}
			wrapped := map[string]json.RawMessage{"data": wrapper.body.Bytes()}
			out, err := json.Marshal(wrapped)
			if err != nil {
				w.WriteHeader(wrapper.statusCode)
				_, _ = w.Write(wrapper.body.Bytes())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(wrapper.statusCode)
			_, _ = w.Write(out)
			return
		}

		w.WriteHeader(wrapper.statusCode)
		_, _ = w.Write(wrapper.body.Bytes())
	})
}
