/*  crypto.go
*
* @Author:             Nanang Suryadi
* @Date:               November 28, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 28/11/19 18:20
 */

package suki

import (
        "encoding/base64"
        "strings"
)

var b64 = base64.RawStdEncoding

// PassLibBase64Encode encodes using a variant of base64, like Passlib.
// Check https://pythonhosted.org/passlib/lib/passlib.utils.html#passlib.utils.ab64_encode
func PassLibBase64Encode(src []byte) (dst string) {
        dst = b64.EncodeToString(src)
        dst = strings.Replace(dst, "+", ".", -1)
        return
}

// PassLibBase64Decode decodes using a variant of base64, like Passlib.
// Check https://pythonhosted.org/passlib/lib/passlib.utils.html#passlib.utils.ab64_decode
func PassLibBase64Decode(src string) (dst []byte, err error) {
        src = strings.Replace(src, ".", "+", -1)
        dst, err = b64.DecodeString(src)
        return
}

// Base64Encode encodes using a Standard of base64.
// return string base64 encode
func Base64Encode(src []byte) (dst string) {
        return base64.StdEncoding.EncodeToString(src)
}

// Base64Encode decodes using a Standard of base64.
// return string base64 encode
func Base64Decode(src string) (dst []byte) {
        decode, err := base64.StdEncoding.DecodeString(src)
        if err != nil {
                panic(err)
        }
        return decode
}
