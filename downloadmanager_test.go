package downloadmanager

import (
	"os"
	"testing"

	"github.com/rai-project/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestDownloadJSON(t *testing.T) {
	const url = "http://data.dmlc.ml/models/imagenet/inception-bn/Inception-BN-symbol.json"
	ctx := context.Background()
	opts := NewOptions(MD5Sum("93ea4544c19709161fec0051aea34885"))
	target, err := DownloadFile(ctx, url, opts.md5Sum)
	assert.NoError(t, err)
	assert.NotEmpty(t, target)
}

func TestMain(m *testing.M) {
	config.Init(
		config.AppName("carml"),
		config.VerboseMode(true),
		config.DebugMode(true),
	)
	os.Exit(m.Run())
}
