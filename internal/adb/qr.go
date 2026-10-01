package adb

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	qrServicePrefix   = "droidlink-"
	qrServiceIDLength = 8
	qrPasswordLength  = 12
	qrAlphabet        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// QRPairing holds the credentials encoded in an Android Wireless Debugging
// "Pair device with QR code" payload. After scanning, the device advertises
// _adb-tls-pairing._tcp under ServiceName and accepts Password as the code.
type QRPairing struct {
	ServiceName string
	Password    string
}

// NewQRPairing generates a random service name and password. Both use a
// plain alphanumeric alphabet so they never need escaping in the payload.
func NewQRPairing() (QRPairing, error) {
	id, err := randomString(qrServiceIDLength)
	if err != nil {
		return QRPairing{}, err
	}

	password, err := randomString(qrPasswordLength)
	if err != nil {
		return QRPairing{}, err
	}

	return QRPairing{ServiceName: qrServicePrefix + id, Password: password}, nil
}

// Payload returns the Wi-Fi style QR string Android's Wireless Debugging
// scanner expects.
func (p QRPairing) Payload() string {
	return fmt.Sprintf("WIFI:T:ADB;S:%s;P:%s;;", p.ServiceName, p.Password)
}

func randomString(length int) (string, error) {
	limit := big.NewInt(int64(len(qrAlphabet)))
	out := make([]byte, length)

	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate random pairing value: %w", err)
		}

		out[i] = qrAlphabet[n.Int64()]
	}

	return string(out), nil
}
