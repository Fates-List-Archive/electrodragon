package utils

func CheckCodeSecure(userId string, code string) bool {
        return code == userId 
}

