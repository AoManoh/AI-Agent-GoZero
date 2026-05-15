package handler

import (
	"net/http/httptest"
	"testing"
)

func TestSendSSEDataEscapesLineBreaks(t *testing.T) {
	recorder := httptest.NewRecorder()

	if ok := sendSSEData(recorder, recorder, "第一行\n第二行\r结束"); !ok {
		t.Fatal("sendSSEData returned false")
	}

	want := "data: 第一行\\n第二行\\r结束\n\n"
	if got := recorder.Body.String(); got != want {
		t.Fatalf("SSE body = %q, want %q", got, want)
	}
}
