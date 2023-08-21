package gsign

import (
	"github.com/donetkit/contrib/utils/grand"
	"testing"
)

func TestSign(t *testing.T) {

	params := make(map[string]string)
	params["appid"] = "appid"
	params["nonce_str"] = grand.RandString(6)
	params["ip"] = "192.168.5.110"
	params["sign_type"] = "MD5"

	t.Log(ParamSign(params, "123456"))
	t.Log("e847f0c7d0b4b95deafed23d235c0a2b")
}
