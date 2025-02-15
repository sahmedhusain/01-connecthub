package image

import "encoding/base64"

func Base64EncodeImage(img []byte) string {
	return base64.StdEncoding.EncodeToString(img)
}
