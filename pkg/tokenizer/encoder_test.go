package tokenizer

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBytesToUnicode(t *testing.T) {
	is := assert.New(t)

	// Most useful test E.V.E.R ^^
	want := map[rune]string{
		0:   "Ā",
		1:   "ā",
		2:   "Ă",
		3:   "ă",
		4:   "Ą",
		5:   "ą",
		6:   "Ć",
		7:   "ć",
		8:   "Ĉ",
		9:   "ĉ",
		10:  "Ċ",
		11:  "ċ",
		12:  "Č",
		13:  "č",
		14:  "Ď",
		15:  "ď",
		16:  "Đ",
		17:  "đ",
		18:  "Ē",
		19:  "ē",
		20:  "Ĕ",
		21:  "ĕ",
		22:  "Ė",
		23:  "ė",
		24:  "Ę",
		25:  "ę",
		26:  "Ě",
		27:  "ě",
		28:  "Ĝ",
		29:  "ĝ",
		30:  "Ğ",
		31:  "ğ",
		32:  "Ġ",
		33:  "!",
		34:  "\"",
		35:  "#",
		36:  "$",
		37:  "%",
		38:  "&",
		39:  "'",
		40:  "(",
		41:  ")",
		42:  "*",
		43:  "+",
		44:  ",",
		45:  "-",
		46:  ".",
		47:  "/",
		48:  "0",
		49:  "1",
		50:  "2",
		51:  "3",
		52:  "4",
		53:  "5",
		54:  "6",
		55:  "7",
		56:  "8",
		57:  "9",
		58:  ":",
		59:  ";",
		60:  "<",
		61:  "=",
		62:  ">",
		63:  "?",
		64:  "@",
		65:  "A",
		66:  "B",
		67:  "C",
		68:  "D",
		69:  "E",
		70:  "F",
		71:  "G",
		72:  "H",
		73:  "I",
		74:  "J",
		75:  "K",
		76:  "L",
		77:  "M",
		78:  "N",
		79:  "O",
		80:  "P",
		81:  "Q",
		82:  "R",
		83:  "S",
		84:  "T",
		85:  "U",
		86:  "V",
		87:  "W",
		88:  "X",
		89:  "Y",
		90:  "Z",
		91:  "[",
		92:  "\\",
		93:  "]",
		94:  "^",
		95:  "_",
		96:  "`",
		97:  "a",
		98:  "b",
		99:  "c",
		100: "d",
		101: "e",
		102: "f",
		103: "g",
		104: "h",
		105: "i",
		106: "j",
		107: "k",
		108: "l",
		109: "m",
		110: "n",
		111: "o",
		112: "p",
		113: "q",
		114: "r",
		115: "s",
		116: "t",
		117: "u",
		118: "v",
		119: "w",
		120: "x",
		121: "y",
		122: "z",
		123: "{",
		124: "|",
		125: "}",
		126: "~",
		127: "ġ",
		128: "Ģ",
		129: "ģ",
		130: "Ĥ",
		131: "ĥ",
		132: "Ħ",
		133: "ħ",
		134: "Ĩ",
		135: "ĩ",
		136: "Ī",
		137: "ī",
		138: "Ĭ",
		139: "ĭ",
		140: "Į",
		141: "į",
		142: "İ",
		143: "ı",
		144: "Ĳ",
		145: "ĳ",
		146: "Ĵ",
		147: "ĵ",
		148: "Ķ",
		149: "ķ",
		150: "ĸ",
		151: "Ĺ",
		152: "ĺ",
		153: "Ļ",
		154: "ļ",
		155: "Ľ",
		156: "ľ",
		157: "Ŀ",
		158: "ŀ",
		159: "Ł",
		160: "ł",
		161: "¡",
		162: "¢",
		163: "£",
		164: "¤",
		165: "¥",
		166: "¦",
		167: "§",
		168: "¨",
		169: "©",
		170: "ª",
		171: "«",
		172: "¬",
		173: "Ń",
		174: "®",
		175: "¯",
		176: "°",
		177: "±",
		178: "²",
		179: "³",
		180: "´",
		181: "µ",
		182: "¶",
		183: "·",
		184: "¸",
		185: "¹",
		186: "º",
		187: "»",
		188: "¼",
		189: "½",
		190: "¾",
		191: "¿",
		192: "À",
		193: "Á",
		194: "Â",
		195: "Ã",
		196: "Ä",
		197: "Å",
		198: "Æ",
		199: "Ç",
		200: "È",
		201: "É",
		202: "Ê",
		203: "Ë",
		204: "Ì",
		205: "Í",
		206: "Î",
		207: "Ï",
		208: "Ð",
		209: "Ñ",
		210: "Ò",
		211: "Ó",
		212: "Ô",
		213: "Õ",
		214: "Ö",
		215: "×",
		216: "Ø",
		217: "Ù",
		218: "Ú",
		219: "Û",
		220: "Ü",
		221: "Ý",
		222: "Þ",
		223: "ß",
		224: "à",
		225: "á",
		226: "â",
		227: "ã",
		228: "ä",
		229: "å",
		230: "æ",
		231: "ç",
		232: "è",
		233: "é",
		234: "ê",
		235: "ë",
		236: "ì",
		237: "í",
		238: "î",
		239: "ï",
		240: "ð",
		241: "ñ",
		242: "ò",
		243: "ó",
		244: "ô",
		245: "õ",
		246: "ö",
		247: "÷",
		248: "ø",
		249: "ù",
		250: "ú",
		251: "û",
		252: "ü",
		253: "ý",
		254: "þ",
		255: "ÿ",
	}

	got := bytesToUnicode()
	is.EqualValues(want, got)
}

func TestNewEncoder_bpeRank(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	is.EqualValues(49830, encoder.bpeRank[lo.T2("c", "rim")])
	is.EqualValues(49880, encoder.bpeRank[lo.T2("Ġdispens", "ary")])
	is.EqualValues(49905, encoder.bpeRank[lo.T2("ĠAm", "p")])
}

func TestNewEncoder_encoder(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	is.EqualValues(50225, encoder.encoder["Ġreclaimed"])
	is.EqualValues(50145, encoder.encoder["headers"])
	is.EqualValues(50256, encoder.encoder["<|endoftext|>"])
}

func TestNewEncoder_decoder(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	is.EqualValues("Ġreclaimed", encoder.decoder[50225])
	is.EqualValues("headers", encoder.decoder[50145])
	is.EqualValues("<|endoftext|>", encoder.decoder[50256])
}

func TestNewEncoder_splitToken(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	want := []string{
		"hello",
		" 👋",
		" world",
		" 🌍",
		" This",
		" is",
		" a",
		" long",
		" string",
		" to",
		" test",
		" whether",
		" or",
		" not",
		" the",
		" emoji",
		" issue",
		" was",
		" fixed",
		"!",
	}

	got, err := encoder.splitToken("hello 👋 world 🌍 This is a long string to test whether or not the emoji issue was fixed!")
	is.EqualValues(want, got)
	is.Nil(err)
}

func TestGetPairs(t *testing.T) {
	is := assert.New(t)

	words := [][]string{
		{"h", "e", "l", "l", "o"},
		{"he", "l", "l", "o"},
		{"hel", "l", "o"},
		{"hell", "o"},
		{"hello"},
		{"Ġ", "ð", "Ł", "ĳ", "ĭ"},
		{"Ġ", "w", "o", "r", "l", "d"},
		{"Ġ", "ð", "Ł", "Į", "į"},
		{"Ġ", "T", "h", "i", "s"},
		{"Ġ", "i", "s"},
		{"Ġ", "a"},
		{"Ġ", "l", "o", "n", "g"},
		{"Ġ", "s", "t", "r", "i", "n", "g"},
		{"Ġ", "t", "o"},
		{"Ġ", "t", "e", "s", "t"},
		{"Ġ", "w", "h", "e", "t", "h", "e", "r"},
		{"Ġ", "o", "r"},
		{"Ġ", "n", "o", "t"},
		{"Ġ", "t", "h", "e"},
		{"Ġ", "e", "m", "o", "j", "i"},
		{"Ġ", "i", "s", "s", "u", "e"},
		{"Ġ", "w", "a", "s"},
		{"Ġ", "f", "i", "x", "e", "d"},
		{"!"},
	}

	wants := [][]lo.Tuple2[string, string]{
		{lo.T2("h", "e"), lo.T2("e", "l"), lo.T2("l", "l"), lo.T2("l", "o")},
		{lo.T2("he", "l"), lo.T2("l", "l"), lo.T2("l", "o")},
		{lo.T2("hel", "l"), lo.T2("l", "o")},
		{lo.T2("hell", "o")},
		{},
		{lo.T2("Ġ", "ð"), lo.T2("ð", "Ł"), lo.T2("Ł", "ĳ"), lo.T2("ĳ", "ĭ")},
		{lo.T2("Ġ", "w"), lo.T2("w", "o"), lo.T2("o", "r"), lo.T2("r", "l"), lo.T2("l", "d")},
		{lo.T2("Ġ", "ð"), lo.T2("ð", "Ł"), lo.T2("Ł", "Į"), lo.T2("Į", "į")},
		{lo.T2("Ġ", "T"), lo.T2("T", "h"), lo.T2("h", "i"), lo.T2("i", "s")},
		{lo.T2("Ġ", "i"), lo.T2("i", "s")},
		{lo.T2("Ġ", "a")},
		{lo.T2("Ġ", "l"), lo.T2("l", "o"), lo.T2("o", "n"), lo.T2("n", "g")},
		{lo.T2("Ġ", "s"), lo.T2("s", "t"), lo.T2("t", "r"), lo.T2("r", "i"), lo.T2("i", "n"), lo.T2("n", "g")},
		{lo.T2("Ġ", "t"), lo.T2("t", "o")},
		{lo.T2("Ġ", "t"), lo.T2("t", "e"), lo.T2("e", "s"), lo.T2("s", "t")},
		{lo.T2("Ġ", "w"), lo.T2("w", "h"), lo.T2("h", "e"), lo.T2("e", "t"), lo.T2("t", "h"), lo.T2("e", "r")},
		{lo.T2("Ġ", "o"), lo.T2("o", "r")},
		{lo.T2("Ġ", "n"), lo.T2("n", "o"), lo.T2("o", "t")},
		{lo.T2("Ġ", "t"), lo.T2("t", "h"), lo.T2("h", "e")},
		{lo.T2("Ġ", "e"), lo.T2("e", "m"), lo.T2("m", "o"), lo.T2("o", "j"), lo.T2("j", "i")},
		{lo.T2("Ġ", "i"), lo.T2("i", "s"), lo.T2("s", "s"), lo.T2("s", "u"), lo.T2("u", "e")},
		{lo.T2("Ġ", "w"), lo.T2("w", "a"), lo.T2("a", "s")},
		{lo.T2("Ġ", "f"), lo.T2("f", "i"), lo.T2("i", "x"), lo.T2("x", "e"), lo.T2("e", "d")},
		{},
	}

	for i := range words {
		want := wants[i]
		got := getPairs(words[i])
		is.EqualValues(want, got, i)
	}

	got := getPairs([]string{"hello"})
	is.EqualValues([]lo.Tuple2[string, string]{}, got)
}

func TestNewEncoder_bpe(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	cases := []lo.Tuple2[string, string]{
		lo.T2("hello", "hello"),
		lo.T2("ĠðŁĳĭ", "ĠðŁĳ ĭ"),
		lo.T2("Ġworld", "Ġworld"),
		lo.T2("ĠðŁĮį", "ĠðŁ Į į"),
		lo.T2("ĠThis", "ĠThis"),
		lo.T2("Ġis", "Ġis"),
		lo.T2("Ġa", "Ġa"),
		lo.T2("Ġlong", "Ġlong"),
		lo.T2("Ġstring", "Ġstring"),
		lo.T2("Ġto", "Ġto"),
		lo.T2("Ġtest", "Ġtest"),
		lo.T2("Ġwhether", "Ġwhether"),
		lo.T2("Ġor", "Ġor"),
		lo.T2("Ġnot", "Ġnot"),
		lo.T2("Ġthe", "Ġthe"),
		lo.T2("Ġemoji", "Ġemoji"),
		lo.T2("Ġissue", "Ġissue"),
		lo.T2("Ġwas", "Ġwas"),
		lo.T2("Ġfixed", "Ġfixed"),
	}

	for _, c := range cases {
		got := encoder.bpe(c.A)
		want := c.B
		is.EqualValues(want, got)
	}
}

func TestNewEncoder_encode(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	want := []int{31373, 995, 770, 318, 257, 890, 4731, 284, 1332, 1771, 393, 407, 262, 44805, 2071, 373, 5969, 0}
	got, err := encoder.Encode("hello world This is a long string to test whether or not the emoji issue was fixed!")
	is.EqualValues(want, got)
	is.Nil(err)

	// @TODO
	// want = []int{31373, 50169, 233, 995, 12520, 234, 235, 770, 318, 257, 890, 4731, 284, 1332, 1771, 393, 407, 262, 44805, 2071, 373, 5969, 0}
	// got, err = encoder.Encode("hello 👋 world 🌍 This is a long string to test whether or not the emoji issue was fixed!")
	// is.EqualValues(want, got)
	// is.Nil(err)
}

func TestNewEncoder_decode(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	want := "hello world This is a long string to test whether or not the emoji issue was fixed!"
	got := encoder.Decode([]int{31373, 995, 770, 318, 257, 890, 4731, 284, 1332, 1771, 393, 407, 262, 44805, 2071, 373, 5969, 0})
	is.EqualValues(want, got)

	// @TODO
	// want = "hello 👋 world 🌍 This is a long string to test whether or not the emoji issue was fixed!"
	// got = encoder.Decode([]int{31373, 50169, 233, 995, 12520, 234, 235, 770, 318, 257, 890, 4731, 284, 1332, 1771, 393, 407, 262, 44805, 2071, 373, 5969, 0})
	// is.EqualValues(want, got)
}

func TestNewEncoder_e2e(t *testing.T) {
	is := assert.New(t)

	encoder, err := NewEncoder()
	is.Nil(err)

	cases := []lo.Tuple2[string, []int]{
		// lo.T2("", []int{}),	// @TODO
		lo.T2(" ", []int{220}),
		lo.T2("\t", []int{197}),
		lo.T2("This is some text", []int{1212, 318, 617, 2420}),
		lo.T2("indivisible", []int{521, 452, 12843}),
		// lo.T2("hello 👋 world 🌍", []int{31373, 50169, 233, 995, 12520, 234, 235}),	// @TODO
	}

	for _, c := range cases {
		encoded, err := encoder.Encode(c.A)
		is.Nil(err)
		is.EqualValues(c.B, encoded, c.A)

		result := encoder.Decode(encoded)
		is.EqualValues(c.A, result, c.A)
	}
}
