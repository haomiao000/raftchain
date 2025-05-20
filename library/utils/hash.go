package utils
import(
	"crypto/sha256"
	"encoding/hex"
)

func GenSha256(str string) string{
	hasher := sha256.New()
	hasher.Write([]byte(str))
	hashStr := hasher.Sum(nil)
	return hex.EncodeToString(hashStr)
}