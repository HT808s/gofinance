package main

import "github.com/fatih/color"

/* prints either green or red text to the screen, depending
 * on decision. */
func binary(text string, decision bool) string {
	if decision {
		return color.GreenString(text)
	}
	return color.RedString(text)
}

func binaryf(f float64, decision bool) string {
	if decision {
		return color.GreenString("%.2f", f)
	}
	return color.RedString("%.2f", f)
}

func binaryfp(f float64, decision bool) string {
	if decision {
		return color.GreenString("%.2f%%", f)
	}
	return color.RedString("%.2f%%", f)
}

// func numberf(f float64) string {
// 	return number("%.2f", f)
// }
//
// func numberfp(f float64) string {
// 	return number("%.2f%%", f)
// }

func arrow(decision bool) string {
	if decision {
		return "↑"
	}
	return "↓"
}
