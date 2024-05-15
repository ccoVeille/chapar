package domain

const (
	ApiVersion = "v1"

	KindConfig        = "Config"
	KindWorkspace     = "Workspace"
	KindEnv           = "Environment"
	KindRequest       = "Request"
	KindPreferences   = "Preferences"
	KindCollection    = "Collection"
	KindProtoFileList = "ProtoFileList"
)

type MetaData struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type KeyValue struct {
	ID     string `yaml:"id"`
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	Enable bool   `yaml:"enable"`
}

// CompareKeyValues compares two slices of KeyValue and returns true if they are equal
func CompareKeyValues(a, b []KeyValue) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !CompareEnvValue(v, b[i]) {
			return false
		}
	}

	return true
}

func KeyValuesToText(values []KeyValue) string {
	var text string
	for _, v := range values {
		text += v.Key + ": " + v.Value + "\n"
	}
	return text
}
