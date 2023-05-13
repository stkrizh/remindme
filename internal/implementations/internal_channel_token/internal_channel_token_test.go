package internalchanneltoken

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHMACInternalChannelTokenGenerator(t *testing.T) {
	cases := []struct {
		secretKey string
	}{
		{secretKey: "test-1"},
		{secretKey: "test-2"},
		{secretKey: "test-3"},
	}

	for _, testCase := range cases {
		t.Run(testCase.secretKey, func(t *testing.T) {
			g := NewHMAC(testCase.secretKey)

			token := g.GenerateInternalChannelToken()
			if !g.ValidateInternalChannelToken(token) {
				require.FailNow(t, "Token invalid.", token)
			}
		})
	}
}
