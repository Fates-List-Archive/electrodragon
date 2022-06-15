package utils

// This is a dummy keygen, the real 'keygen_secure.go' is not public
// For self hosting, you must make your own keygen for performing staff verification
// Your secure keygen should have a function called `checkCodeSecure` that takes
// a user ID and a code and returns true if the code is valid for the user
func checkCodeDev(userId string, code string) bool {
	return true // In dev mode, always return true
}
