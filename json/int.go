package json

func appendInt(b []byte, n int64) []byte {
	return formatInteger(b, uint64(n), n < 0)
}

func appendUint(b []byte, n uint64) []byte {
	return formatInteger(b, n, false)
}

const intLookup = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

func formatInteger(b []byte, n uint64, negative bool) []byte {
	if !negative {
		if n < 10 {
			return append(b, byte(n+'0'))
		} else if n < 100 {
			i := int(n)
			return append(b, intLookup[i*2:i*2+2]...)
		}
	} else {
		n = -n
	}

	var a [20 + 1]byte // sign + up to 20 digits for UINT64_MAX
	i := len(a)

	for n >= 100 {
		is := n % 100 * 2
		n /= 100
		i -= 2
		a[i+1] = intLookup[is+1]
		a[i+0] = intLookup[is+0]
	}

	is := n * 2
	i--
	a[i] = intLookup[is+1]
	if n >= 10 {
		i--
		a[i] = intLookup[is]
	}

	if negative {
		i--
		a[i] = '-'
	}

	return append(b, a[i:]...)
}
