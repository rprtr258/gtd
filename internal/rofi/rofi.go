package rofi

import (
	"fmt"
	"os"
)

// IsFirstOpen check if rofi menu was first opened and
// no variant is selected yet.
func IsFirstOpen() bool {
	return os.Getenv("ROFI_RETV") == "0"
}

// YieldItemWithInfo prints menu item with info which can be
// retrieved later with GetInfo
func YieldItemWithInfo(text string, info string) {
	fmt.Printf("%s\x00info\x1f%s\n", text, info)
}

// GetInfo from chosen menu item
func GetInfo() string {
	return os.Getenv("ROFI_INFO")
}
