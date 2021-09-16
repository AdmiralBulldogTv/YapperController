package jwt

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/troydota/tts-textparser/src/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var alg = jwt.EncodeSegment(utils.S2B(`{"alg":"HS256","typ":"JWT"}`))

func Sign(secret string, pl interface{}) (string, error) {
	bytes, err := json.MarshalToString(pl)
	if err != nil {
		return "", err
	}

	first := fmt.Sprintf("%s.%s", alg, jwt.EncodeSegment(utils.S2B(bytes)))
	sign, err := jwt.SigningMethodHS256.Sign(first, utils.S2B(secret))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", first, sign), nil
}

func Verify(secret string, token string, out interface{}) error {
	tokenSplits := strings.Split(token, ".")
	if len(tokenSplits) != 3 {
		return jwt.ErrInvalidKey
	}

	if err := jwt.SigningMethodHS256.Verify(fmt.Sprintf("%s.%s", tokenSplits[0], tokenSplits[1]), tokenSplits[2], utils.S2B(secret)); err != nil {
		return err
	}

	val, err := jwt.DecodeSegment(tokenSplits[1])
	if err != nil {
		return err
	}

	return json.Unmarshal(val, out)
}
