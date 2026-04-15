package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// IsNIF valide un NIF algérien (15 chiffres)
func IsNIF(nif string) bool {
	if len(nif) == 0 {
		return true // Optionnel
	}
	nif = strings.ReplaceAll(nif, " ", "")
	if len(nif) != 15 {
		return false
	}
	for _, c := range nif {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsNIS valide un NIS algérien (11 chiffres)
func IsNIS(nis string) bool {
	if len(nis) == 0 {
		return true
	}
	nis = strings.ReplaceAll(nis, " ", "")
	if len(nis) != 11 {
		return false
	}
	for _, c := range nis {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsRC valide un registre de commerce algérien
func IsRC(rc string) bool {
	if len(rc) == 0 {
		return true
	}
	// Format: WW/MM-NNNNNNN L XX (simplifié)
	return len(rc) >= 5
}

// IsRIB valide un RIB algérien (20 chiffres)
func IsRIB(rib string) bool {
	if len(rib) == 0 {
		return true
	}
	rib = strings.ReplaceAll(rib, " ", "")
	if len(rib) != 20 {
		return false
	}
	for _, c := range rib {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsEmail valide une adresse email
func IsEmail(email string) bool {
	if email == "" {
		return true
	}
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsPhoneAlgerian valide un numéro de téléphone algérien
func IsPhoneAlgerian(phone string) bool {
	if phone == "" {
		return true
	}
	phone = strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	phone = strings.ReplaceAll(phone, ".", "")
	if strings.HasPrefix(phone, "+213") {
		phone = "0" + phone[4:]
	}
	if len(phone) != 10 {
		return false
	}
	for _, c := range phone {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsMobileAlgerian valide un numéro mobile algérien (05/06/07)
func IsMobileAlgerian(mobile string) bool {
	if mobile == "" {
		return true
	}
	mobile = strings.ReplaceAll(strings.ReplaceAll(mobile, " ", ""), "-", "")
	if !strings.HasPrefix(mobile, "05") &&
		!strings.HasPrefix(mobile, "06") &&
		!strings.HasPrefix(mobile, "07") {
		return false
	}
	return IsPhoneAlgerian(mobile)
}

// IsPositiveNumber valide un nombre positif
func IsPositiveNumber(s string) bool {
	if s == "" {
		return false
	}
	val, err := strconv.ParseFloat(s, 64)
	return err == nil && val >= 0
}

// IsPositiveInt valide un entier positif
func IsPositiveInt(s string) bool {
	if s == "" {
		return false
	}
	val, err := strconv.Atoi(s)
	return err == nil && val >= 0
}

// IsPercentage valide un pourcentage (0-100)
func IsPercentage(s string) bool {
	if s == "" {
		return false
	}
	val, err := strconv.ParseFloat(s, 64)
	return err == nil && val >= 0 && val <= 100
}

// IsEAN13 valide un code-barres EAN-13
func IsEAN13(barcode string) bool {
	if len(barcode) != 13 {
		return false
	}
	for _, c := range barcode {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	// Vérification du chiffre de contrôle
	sum := 0
	for i, c := range barcode[:12] {
		digit := int(c - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	return checkDigit == int(barcode[12]-'0')
}

// GenerateEAN13CheckDigit génère le chiffre de contrôle EAN-13
func GenerateEAN13CheckDigit(barcode12 string) string {
	if len(barcode12) != 12 {
		return ""
	}
	sum := 0
	for i, c := range barcode12 {
		digit := int(c - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	return barcode12 + strconv.Itoa(checkDigit)
}

// IsZipCode valide un code postal (5 chiffres)
func IsZipCode(zip string) bool {
	if zip == "" {
		return true
	}
	zip = strings.TrimSpace(zip)
	if len(zip) != 5 {
		return false
	}
	for _, c := range zip {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// IsURL valide une URL simplifiée
func IsURL(url string) bool {
	if url == "" {
		return true
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "www.")
}

// RequiredNotEmpty vérifie qu'un champ n'est pas vide
func RequiredNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// MinLength vérifie la longueur minimale
func MinLength(s string, min int) bool {
	return len([]rune(strings.TrimSpace(s))) >= min
}
