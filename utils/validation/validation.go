package validation

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"unicode"
)

func ValidateEmployeeNameToDiscord(employeeFullName string) (string, error) {
	employeeFullNameSplited := strings.Split(employeeFullName, " ")
	if len(employeeFullNameSplited) != 3 {
		return "", fmt.Errorf("Неверный формат имени работника - '%s'", employeeFullName)
	}

	discordEmployeeName := employeeFullNameSplited[:2]

	for _, name := range discordEmployeeName {
		for _, char := range name {
			if !unicode.Is(unicode.Cyrillic, char) {
				return "", fmt.Errorf("Имя работника должно быть прописано кириллицей - '%s'", employeeFullName)
			}
		}
	}

	return cases.Title(language.Russian).String(strings.Join(discordEmployeeName, " ")), nil
}
