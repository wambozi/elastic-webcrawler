package conf

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func trimStringFromSubstring(s string, substring string) string {
	idx := strings.Index(s, substring)
	if idx != -1 {
		return s[:idx]
	}
	return s
}

func TestSetup(t *testing.T) {
	type results struct {
		Conf   *Configuration
		ErrMsg string
	}

	tests := map[string]struct {
		env    string
		conf   *Configuration
		errMsg string
	}{
		"test": {
			env: "test",
			conf: &Configuration{Server: ServerConfiguration{Port: 8081, ReadHeaderTimeoutMillis: 3000},
				Appsearch: AppsearchOptions{
					Endpoint: "http://localhost:3002",
					API:      "/api/as/v1/",
					Token:    "private-somefakek3y",
				},
				Elasticsearch: ElasticOptions{
					Endpoint: "http://localhost:9200",
					Username: "elastic",
					Password: "changeme",
				},
			}, errMsg: ""},
		"incorrect env": {env: "other", conf: nil, errMsg: "Error reading config file. env: other error: Config File \"other\" Not"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotConfs, gotErrs := Setup(tc.env)

			var gotErrMsg string
			//this needs to be updated if there are additional environment variables used in the future
			if gotErrs != nil {
				gotErrMsg = gotErrs.Error()
			}

			//guarding to avoid nil pointer dereferences below
			if tc.conf == nil && gotConfs != nil {
				confDiffMsg := fmt.Sprintf("configuration | expected : %+v | received : %+v", tc.conf, gotConfs)
				t.Fatal(confDiffMsg)
			}

			trimSubstring := " Found in"
			expectedErrMsgTrim := trimStringFromSubstring(tc.errMsg, trimSubstring)
			gotErrMsgTrim := trimStringFromSubstring(gotErrMsg, trimSubstring)

			want := results{tc.conf, expectedErrMsgTrim}
			got := results{gotConfs, gotErrMsgTrim}
			diff := cmp.Diff(want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := make(map[string]string)
	tests[""] = "<nil>"
	tests["lle"] = "lle"
	tests["prod"] = "prod"
	tests["someOtherValue"] = "someOtherValue"

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			os.Setenv("ENV_ID", k)
			got := GetEnvironment()
			diff := cmp.Diff(v, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}

}
