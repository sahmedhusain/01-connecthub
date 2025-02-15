package image

import "encoding/base64"

func base64EncodeImage(img []byte) string {
	return base64.StdEncoding.EncodeToString(img)
}
