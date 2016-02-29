package command

import "testing"

func TestTriggerTestArgsValidity(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: "a valid triggerTestArgs object",
			in: &triggerTestArgs{
				projectId:   1,
				triggerName: "trigger",
				parameters: setFlags{
					"key,value": {},
				},
			},
			want: true,
		},
		{
			desc: "a valid triggerTestArgs object w/ no parameters",
			in: &triggerTestArgs{
				projectId:   1,
				triggerName: "trigger",
			},
			want: true,
		},
		{
			desc: testDescInvalidProjectId,
			in: &triggerTestArgs{
				projectId:   0,
				triggerName: "trigger",
				parameters: setFlags{
					"key,value": {},
				},
			},
			want: false,
		},
		{
			desc: "invalid trigger name (none)",
			in: &triggerTestArgs{
				projectId: 1,
				parameters: setFlags{
					"key,value": {},
				},
			},
			want: false,
		},
		{
			desc: "invalid parameter (not comma separated)",
			in: &triggerTestArgs{
				projectId:   1,
				triggerName: "trigger",
				parameters: setFlags{
					"key value": {},
				},
			},
			want: false,
		},
	}

	runDataTestCase(t, cases)
}

func TestTriggerDataValidity(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: testDescInvalidProjectId,
			in: &triggerData{
				TriggerId:   0,
				ProjectId:   0, // must be > 0
				TriggerName: "trigger",
				DataExpiry:  0,
			},
			want: false,
		},
		{
			desc: "invalid trigger name (none)",
			in: &triggerData{
				TriggerId:  0,
				ProjectId:  1,
				DataExpiry: 0,
			},
			want: false,
		},
		{
			desc: "valid triggerData object",
			in: &triggerData{
				TriggerId:   0,
				ProjectId:   1,
				TriggerName: "trigger",
				DataExpiry:  0,
			},
			want: true,
		},
	}

	runDataTestCase(t, cases)
}

func TestHTTPDataValidity(t *testing.T) {
	cases := []struct {
		in   *httpData
		want bool
	}{
		{
			in: &httpData{
				URL:         "", // must have len > 0
				ContentType: "text/plain",
			},
			want: false,
		},
		{
			in: &httpData{
				URL:         "iobeam.com",
				ContentType: "",
			},
			want: false,
		},
		{
			in: &httpData{
				URL:         "iobeam.com",
				ContentType: "text/plain",
			},
			want: true,
		},
	}

	for _, c := range cases {
		if got := c.in.isHTTPDataValid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMQTTDataValidity(t *testing.T) {
	cases := []struct {
		in   *mqttData
		want bool
	}{
		{
			in: &mqttData{
				Broker:  "iobeam.com",
				Topic:   "good topic",
				Payload: "message",
			},
			want: true,
		},
		{
			in: &mqttData{
				Broker:  "", // must have len > 0
				Topic:   "good topic",
				Payload: "message",
			},
			want: false,
		},
		{
			in: &mqttData{
				Broker:  "iobeam.com",
				Topic:   "", // must have len > 0
				Payload: "message",
			},
			want: false,
		},
		{
			in: &mqttData{
				Broker:  "iobeam.com",
				Topic:   "good topic",
				Payload: "", // must have len > 0
			},
			want: false,
		},
	}

	for _, c := range cases {
		if got := c.in.isMQTTDataValid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSMSDataValidity(t *testing.T) {
	cases := []struct {
		in   *smsData
		want bool
	}{
		{
			in: &smsData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: true,
		},
		{
			in: &smsData{
				AccountSID: "", // must have len > 0
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsData{
				AccountSID: "my id",
				AuthToken:  "", // must have len > 0
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "", // must have len > 0
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "", // must have len > 0
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "", // must have len > 0
			},
			want: false,
		},
	}

	for _, c := range cases {
		if got := c.in.isSMSDataValid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
