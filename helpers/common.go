package helpers

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"crypto/md5"
	"encoding/hex"
)

type Common struct {

}

func (c Common) DateTime() string  {

	var t int64 = time.Now().Unix()
	var s string = time.Unix(t, 0).Format("2006-01-02 15:04:05")
	return s
}

func (c Common) PasswrodEncode(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return  "", err
	}
	return  string(hash), nil

}

func (c Common) Md5String(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func (c Common) addError(name string, value string)  {

}

