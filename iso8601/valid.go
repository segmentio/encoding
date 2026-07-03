package iso8601

// ValidFlags is a bitset type used to configure the behavior of the Valid
// function.
type ValidFlags int

const (
	// Strict is a validation flag used to represent a string iso8601 validation
	// (this is the default).
	Strict ValidFlags = 0

	// AllowSpaceSeparator allows the presence of a space instead of a 'T' as
	// separator between the date and time.
	AllowSpaceSeparator ValidFlags = 1 << iota

	// AllowMissingTime allows the value to contain only a date.
	AllowMissingTime

	// AllowMissingSubsecond allows the value to contain only a date and time.
	AllowMissingSubsecond

	// AllowMissingTimezone allows the value to be missing the timezone
	// information.
	AllowMissingTimezone

	// AllowNumericTimezone allows the value to represent timezones in their
	// numeric form.
	AllowNumericTimezone

	// Flexible is a combination of all validation flag that allow for
	// non-strict checking of the input value.
	Flexible = AllowSpaceSeparator | AllowMissingTime | AllowMissingSubsecond | AllowMissingTimezone | AllowNumericTimezone
)

// Valid check value to verify whether or not it is a valid iso8601 time
// representation.
func Valid(value string, flags ValidFlags) bool {
	if flags == Flexible {
		return validFlexible(value)
	}

	var ok bool

	// year
	if value, ok = readDigits(value, 4, 4); !ok {
		return false
	}

	if value, ok = readByte(value, '-'); !ok {
		return false
	}

	// month
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	if value, ok = readByte(value, '-'); !ok {
		return false
	}

	// day
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	if len(value) == 0 && (flags&AllowMissingTime) != 0 {
		return true // date only
	}

	// separator
	if value, ok = readByte(value, 'T'); !ok {
		if (flags & AllowSpaceSeparator) == 0 {
			return false
		}
		if value, ok = readByte(value, ' '); !ok {
			return false
		}
	}

	// hour
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	if value, ok = readByte(value, ':'); !ok {
		return false
	}

	// minute
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	if value, ok = readByte(value, ':'); !ok {
		return false
	}

	// second
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	// microsecond
	if value, ok = readByte(value, '.'); !ok {
		if (flags & AllowMissingSubsecond) == 0 {
			return false
		}
	} else {
		if value, ok = readDigits(value, 1, 9); !ok {
			return false
		}
	}

	if len(value) == 0 && (flags&AllowMissingTimezone) != 0 {
		return true // date and time
	}

	// timezone
	if value, ok = readByte(value, 'Z'); ok {
		return len(value) == 0
	}

	if (flags & AllowSpaceSeparator) != 0 {
		value, _ = readByte(value, ' ')
	}

	if value, ok = readByte(value, '+'); !ok {
		if value, ok = readByte(value, '-'); !ok {
			return false
		}
	}

	// timezone hour
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	if value, ok = readByte(value, ':'); !ok {
		if (flags & AllowNumericTimezone) == 0 {
			return false
		}
	}

	// timezone minute
	if value, ok = readDigits(value, 2, 2); !ok {
		return false
	}

	return len(value) == 0
}

func readDigits(value string, min, max int) (string, bool) {
	if len(value) < min {
		return value, false
	}

	i := 0

	for i < max && i < len(value) && isDigit(value[i]) {
		i++
	}

	if i < max && i < min {
		return value, false
	}

	return value[i:], true
}

func readByte(value string, c byte) (string, bool) {
	if len(value) == 0 {
		return value, false
	}
	if value[0] != c {
		return value, false
	}
	return value[1:], true
}

func validFlexible(value string) bool {
	n := len(value)
	if n < 10 {
		return false
	}

	if !isDigit(value[0]) ||
		!isDigit(value[1]) ||
		!isDigit(value[2]) ||
		!isDigit(value[3]) ||
		value[4] != '-' ||
		!isDigit(value[5]) ||
		!isDigit(value[6]) ||
		value[7] != '-' ||
		!isDigit(value[8]) ||
		!isDigit(value[9]) {
		return false
	}

	if n == 10 {
		return true // date only
	}

	c := value[10]
	if c != 'T' && c != ' ' {
		return false
	}

	if n < 19 {
		return false
	}
	if !isDigit(value[11]) ||
		!isDigit(value[12]) ||
		value[13] != ':' ||
		!isDigit(value[14]) ||
		!isDigit(value[15]) ||
		value[16] != ':' ||
		!isDigit(value[17]) ||
		!isDigit(value[18]) {
		return false
	}

	i := 19
	if i == n {
		return true // missing subsecond and timezone
	}

	if value[i] == '.' {
		i++
		j := i
		end := i + 9
		if end > n {
			end = n
		}
		for i < end && isDigit(value[i]) {
			i++
		}
		if i == j {
			return false
		}
		if i == n {
			return true // missing timezone
		}
	}

	if value[i] == 'Z' {
		return i+1 == n
	}

	if value[i] == ' ' {
		i++
		if i == n {
			return false
		}
	}

	if value[i] != '+' && value[i] != '-' {
		return false
	}
	i++

	if n-i < 4 {
		return false
	}
	if !isDigit(value[i]) || !isDigit(value[i+1]) {
		return false
	}
	i += 2

	if i < n && value[i] == ':' {
		i++
	}

	if n-i != 2 {
		return false
	}
	return isDigit(value[i]) && isDigit(value[i+1])
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
