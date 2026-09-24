package app_common

import (
	"fmt"
	"strings"
)

// The width of the rules above and below a banner. Kept short so that it fits a narrow terminal.
const bannerRuleWidth = 60

// PrintBanner prints a block of text between two rules.
//
// Used for messages that need to stand out from tabular output, such as an account problem that
// explains every row at once. The rules have no sides, so the text can be any length and the terminal
// can wrap it as it sees fit.
func PrintBanner(lines []string) {
	rule := strings.Repeat("#", bannerRuleWidth)

	fmt.Println()
	fmt.Println(rule)

	for _, line := range lines {
		fmt.Println(line)
	}

	fmt.Println(rule)
	fmt.Println()
}
