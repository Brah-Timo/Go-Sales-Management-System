package utils

import (
	"fmt"
	"math"
	"strings"
)

// NumberToWordsFr convertit un montant en lettres (français)
func NumberToWordsFr(amount float64) string {
	amount = math.Round(amount*100) / 100
	intPart := int64(amount)
	centPart := int64(math.Round((amount - float64(intPart)) * 100))

	words := integerToWordsFr(intPart)

	result := strings.TrimSpace(words)
	if intPart > 1 {
		result += " dinars algériens"
	} else {
		result += " dinar algérien"
	}

	if centPart > 0 {
		centWords := integerToWordsFr(centPart)
		if centPart > 1 {
			result += " et " + strings.TrimSpace(centWords) + " centimes"
		} else {
			result += " et " + strings.TrimSpace(centWords) + " centime"
		}
	}

	// Première lettre en majuscule
	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}

	return result
}

var unitsFr = []string{
	"", "un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf",
	"dix", "onze", "douze", "treize", "quatorze", "quinze", "seize",
	"dix-sept", "dix-huit", "dix-neuf",
}

var tensFr = []string{
	"", "", "vingt", "trente", "quarante", "cinquante",
	"soixante", "soixante", "quatre-vingt", "quatre-vingt",
}

func integerToWordsFr(n int64) string {
	if n == 0 {
		return "zéro"
	}
	if n < 0 {
		return "moins " + integerToWordsFr(-n)
	}

	switch {
	case n < 20:
		return unitsFr[n]
	case n < 100:
		return tensWordsFr(n)
	case n < 1000:
		return hundredsWordsFr(n)
	case n < 1_000_000:
		return milleFr(n)
	case n < 1_000_000_000:
		return millionFr(n)
	case n < 1_000_000_000_000:
		return milliardFr(n)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func tensWordsFr(n int64) string {
	if n < 20 {
		return unitsFr[n]
	}
	tens := n / 10
	unit := n % 10

	switch tens {
	case 7:
		// 70-79 → soixante-dix à soixante-dix-neuf
		return "soixante-" + unitsFr[10+unit]
	case 9:
		// 90-99 → quatre-vingt-dix à quatre-vingt-dix-neuf
		return "quatre-vingt-" + unitsFr[10+unit]
	case 8:
		if unit == 0 {
			return "quatre-vingts"
		}
		return "quatre-vingt-" + unitsFr[unit]
	default:
		if unit == 0 {
			return tensFr[tens]
		}
		sep := "-"
		if unit == 1 && tens != 8 {
			sep = "-et-"
		}
		return tensFr[tens] + sep + unitsFr[unit]
	}
}

func hundredsWordsFr(n int64) string {
	hundreds := n / 100
	rest := n % 100

	var centWord string
	if hundreds == 1 {
		centWord = "cent"
	} else {
		centWord = unitsFr[hundreds] + " cent"
		if rest == 0 {
			centWord += "s"
		}
	}

	if rest == 0 {
		return centWord
	}
	return centWord + " " + tensWordsFr(rest)
}

func milleFr(n int64) string {
	thousands := n / 1000
	rest := n % 1000

	var milleWord string
	if thousands == 1 {
		milleWord = "mille"
	} else {
		milleWord = hundredsWordsFr(thousands) + " mille"
	}

	if rest == 0 {
		return milleWord
	}
	return milleWord + " " + hundredsWordsFr(rest)
}

func millionFr(n int64) string {
	millions := n / 1_000_000
	rest := n % 1_000_000

	var millionWord string
	if millions == 1 {
		millionWord = "un million"
	} else {
		millionWord = hundredsWordsFr(millions) + " millions"
	}

	if rest == 0 {
		return millionWord
	}
	return millionWord + " " + milleFr(rest)
}

func milliardFr(n int64) string {
	milliards := n / 1_000_000_000
	rest := n % 1_000_000_000

	var milliardWord string
	if milliards == 1 {
		milliardWord = "un milliard"
	} else {
		milliardWord = hundredsWordsFr(milliards) + " milliards"
	}

	if rest == 0 {
		return milliardWord
	}
	return milliardWord + " " + millionFr(rest)
}
