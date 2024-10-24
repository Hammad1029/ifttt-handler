package resolvable

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"ifttt/handler/common"

	"golang.org/x/crypto/bcrypt"
)

type encodeResolvable struct {
	Input Resolvable `json:"input" mapstructure:"input"`
	Alg   string     `json:"alg" mapstructure:"alg"`
}

func (e *encodeResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	inputResolved, err := e.Input.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	}
	inputStringified := fmt.Sprint(inputResolved)
	inputBArr := []byte(inputStringified)

	switch e.Alg {
	case "md5":
		{
			hash := md5.Sum(inputBArr)
			return hex.EncodeToString(hash[:]), nil
		}
	case "sha1":
		{
			hasher := sha1.New()
			hasher.Write(inputBArr)
			return hex.EncodeToString(hasher.Sum(nil)), nil
		}
	case "sha2":
		{
			hasher := sha256.New()
			hasher.Write(inputBArr)
			return hex.EncodeToString(hasher.Sum(nil)), nil
		}
	case "bcrypt":
		{
			hash, err := bcrypt.GenerateFromPassword(inputBArr, bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			return hex.EncodeToString(hash), nil
		}
	case "base64-de":
		{
			decoded, err := base64.StdEncoding.DecodeString(inputStringified)
			if err != nil {
				return nil, err
			}
			return hex.EncodeToString(decoded), nil
		}
	case "base64-en":
		{
			encoded := base64.StdEncoding.EncodeToString(inputBArr)
			return encoded, nil
		}
	default:
		{
			return nil, fmt.Errorf("encoder for %s not found", e.Alg)
		}
	}
}
