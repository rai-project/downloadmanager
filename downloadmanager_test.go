package downloadmanager

import (
	"os"
	"testing"

	"github.com/rai-project/config"
	"github.com/stretchr/testify/assert"
)

func TestDownloadJSON(t *testing.T) {
	const url = "http://data.dmlc.ml/models/imagenet/inception-bn/Inception-BN-symbol.json"
	target, err := DownloadFile(url, "/tmp/inception.json")
	assert.NoError(t, err)
	assert.NotEmpty(t, target)
	assert.Equal(t, "/tmp/inception.json", target)
}

func TestMain(m *testing.M) {
	config.Init(
		config.AppName("carml"),
		config.VerboseMode(true),
		config.DebugMode(true),
	)
	os.Exit(m.Run())
}
