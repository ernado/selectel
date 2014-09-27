package integration

import (
	"github.com/ernado/selectel/storage"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var (
	ENV_USER = "SELECTEL_USER"
	ENV_KEY  = "SELECTEL_KEY"
)

func TestAuth(t *testing.T) {
	var (
		user, key string
	)
	user = os.Getenv(ENV_USER)
	key = os.Getenv(ENV_KEY)
	Convey("Auth", t, func() {
		c, err := storage.New(user, key)
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
	})
}
