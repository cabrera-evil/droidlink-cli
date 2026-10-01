package ui

import (
	"io"

	"github.com/mdp/qrterminal/v3"
)

const qrQuietZone = 2

// RenderQR writes text as a half-block QR code sized for terminal display.
func RenderQR(w io.Writer, text string) {
	qrterminal.GenerateWithConfig(text, qrterminal.Config{
		Level:          qrterminal.L,
		Writer:         w,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		WhiteChar:      qrterminal.WHITE_WHITE,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		QuietZone:      qrQuietZone,
	})
}
