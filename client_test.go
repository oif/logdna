package logdna

import "testing"

const (
	testApp      = "logdna-go"
	testAPIKey   = "233"
	testHostname = "LogDNA"
)

func TestMustInit(t *testing.T) {
	MustInit(Config{
		App:      testApp,
		APIKey:   testAPIKey,
		Hostname: testHostname,
	})
}

func TestPayload(t *testing.T) {
	p := new(payload)
	p.Write(Line{
		Line: "test",
	})
	if s := p.Size(); s != 1 {
		t.Fatalf("Got unexpected payload size: %d, expect 1", s)
	}
	p.Flush()
	if s := p.Size(); s != 0 {
		t.Fatalf("Got unexpected payload size: %d, expect 0", s)
	}
}

func TestClient(t *testing.T) {
	client := MustInit(Config{
		App:      testApp,
		APIKey:   testAPIKey,
		Hostname: testHostname,
	})
	defer func() {
		err := client.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	client.Write("try to send log")
	client.WriteLine(Line{
		Line: "log body",
	})
	if client.payload.Lines[1].App != client.payload.Lines[0].App || client.payload.Lines[1].App != testApp {
		t.Fatal("Line autocomplete error")
	}

}
