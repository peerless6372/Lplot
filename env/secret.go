package env

import (
	"errors"
	"github.com/peerless6372/Lplot/utils"
	"reflect"
	"strings"
)

func getSecret(prefix string) *utils.Conf {
	fileName := strings.TrimRight(strings.TrimLeft(prefix, "@@"), ".")
	secretConf, err := utils.Load(GetConfDirPath()+"/secret/"+fileName+".secret.yaml", nil)
	if err != nil {
		return nil
	}

	return secretConf
}

// 目前只支持替换为string类型的value
func CommonSecretChange(prefix string, src, dst interface{}) {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)
	if srcType.Kind() != reflect.Struct || dstType.Kind() != reflect.Ptr {
		return
	}

	secretConf := getSecret(prefix)
	if secretConf == nil {
		return
	}

	// 给dst赋值
	val := reflect.ValueOf(dst).Elem()
	for i := 0; i < srcType.NumField(); i++ {
		switch val.Field(i).Kind() {
		case reflect.Struct:
			for j := 0; j < val.Field(i).NumField(); j++ {
				field := val.Field(i).Field(j)
				if rule, ok := getRule(prefix, field); ok {
					n := secretConf.GetString(rule)
					if val.Field(i).Field(j).CanSet() {
						val.Field(i).Field(j).SetString(n)
					}
				}
			}
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			for j := 0; j < val.Field(i).Len(); j++ {
				field := val.Field(i).Index(j)
				if rule, ok := getRule(prefix, field); ok {
					n := secretConf.GetString(rule)
					if val.Field(i).Index(j).CanSet() {
						val.Field(i).Index(j).SetString(n)
					}
				}
			}
		case reflect.String:
			field := val.Field(i)
			if rule, ok := getRule(prefix, field); ok {
				n := secretConf.GetString(rule)
				if val.Field(i).CanSet() {
					val.Field(i).SetString(n)
				}
			}
		}
	}
}

func getRule(prefix string, field reflect.Value) (rule string, ok bool) {
	if field.Kind() != reflect.String {
		return rule, false
	}

	rule = field.String()
	if !strings.HasPrefix(rule, prefix) {
		return rule, false
	}

	// 需要替换的secret key只能是string类型
	rule = rule[len(prefix):]
	return rule, true
}

// 数据库敏感信息加密
const (
	appPrefix           = "app"
	dbEncryptConf       = "rc4"
	dbEncryptConfKey    = "key"
	dbEncryptConfPrefix = "prefix"
)

var dbS *dbSecret

type dbSecret struct {
	key       string
	prefix    string
	prefixLen int
}

func initDBSecret() {
	secretConf := getSecret(appPrefix)
	if secretConf == nil {
		return
	}

	c := secretConf.GetStringMap(dbEncryptConf)
	if len(c) == 0 {
		return
	}

	var key string
	if v, exit := c[dbEncryptConfKey]; exit {
		key, _ = v.(string)
	}

	k := len(key)
	if k < 1 || k > 256 {
		panic("secret/app.secret.yaml has invalid rc4 key, len must [1,256]")
	}

	var prefix string
	if p, exist := c[dbEncryptConfPrefix]; exist && p != nil {
		if p, ok := p.(string); !ok {
			panic("prefix must be string")
		} else {
			prefix = p
		}
	}

	dbS = &dbSecret{
		key:       key,
		prefix:    prefix,
		prefixLen: len(prefix),
	}
}

func EncodeDBSensitiveField(plainText string) string {
	if dbS == nil {
		return plainText
	}
	result, _ := utils.Rc4Encode(dbS.key, plainText)
	return dbS.prefix + result
}

func DecodeDBSensitiveField(encrypted string) string {
	if dbS == nil {
		return encrypted
	}

	if dbS.prefixLen > 0 {
		// 去除前缀再解密
		if len(encrypted) <= dbS.prefixLen {
			return encrypted
		}
		encrypted = encrypted[dbS.prefixLen:]
	}

	result, _ := utils.Rc4Decode(dbS.key, encrypted)
	return result
}

func IsEncrypted(data string) (isEncrypted bool, err error) {
	// 不能识别
	if dbS.prefixLen == 0 {
		return false, errors.New("unrecognized")
	}

	// 配置了前缀，但是字符串长度不足，肯定未加密
	if len(data) <= dbS.prefixLen {
		return false, nil
	}

	// 前缀匹配，认为这个字符串是加密后的
	if data[0:dbS.prefixLen] == dbS.prefix {
		return true, nil
	}

	return false, nil
}
