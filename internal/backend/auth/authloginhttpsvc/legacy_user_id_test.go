package authloginhttpsvc

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestLegacyUserID(t *testing.T) {
	uid := newLegacyUserIDFromUserData(&loginData{ProviderName: "descope", Email: "itay@autokitteh.com", DisplayName: "Itay Donanhirsh"})
	assert.Equal(t, uid.String(), "usr_7s1fretr3kbg7vq9h74k8krkb9")

	uid = newLegacyUserIDFromUserData(&loginData{ProviderName: "descope", Email: "efi@autokitteh.com", DisplayName: "Efi Shtain"})
	assert.Equal(t, uid.String(), "usr_1yd465k7k8dbv130pqv7e3j2xk")
}
