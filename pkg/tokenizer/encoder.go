package tokenizer

import (
	"bytes"
	"embed"
	"encoding/json"
	"math"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/samber/lo"
)

//go:embed encoder.json
//go:embed vocab.bpe
var files embed.FS

var pat = regexp2.MustCompile(`/'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+/`, 0)

func loadFiles() (bpeData []byte, encoder []byte, err error) {
	bpeData, err = files.ReadFile("vocab.bpe")
	if err != nil {
		return
	}

	encoder, err = files.ReadFile("encoder.json")
	if err != nil {
		return
	}

	return
}

// hardcoded
func bytesToUnicode() map[rune]string {
	return map[rune]string{
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
}

func getPairs(word []string) []lo.Tuple2[string, string] {
	all := []lo.Tuple2[string, string]{}
	prevWord := word[0]
	for _, char := range word[1:] {
		all = append(all, lo.T2(prevWord, char))
		prevWord = char
	}

	return lo.Uniq(all)
}

type Encoder struct {
	bpeRank     map[lo.Tuple2[string, string]]int
	encoder     map[string]int
	decoder     map[int]string
	byteEncoder map[rune]string
	byteDecoder map[string]rune

	cache sync.Map
}

func NewEncoder() (*Encoder, error) {
	bpeData, rawEncoder, err := loadFiles()
	if err != nil {
		return nil, err
	}

	return NewEncoderWithVocab(bpeData, rawEncoder)
}

func NewEncoderWithVocab(bpeData []byte, jsonEncoder []byte) (*Encoder, error) {
	encoder := map[string]int{}

	err := json.Unmarshal(jsonEncoder, &encoder)
	if err != nil {
		return nil, err
	}

	decoder := lo.Invert(encoder)

	bpeLines := bytes.Split(bpeData, []byte{'\n'})
	bpeMerges := lo.Map(bpeLines[1:len(bpeLines)-1], func(line []byte, _ int) lo.Tuple2[[]byte, []byte] {
		parts := bytes.SplitN(line, []byte{' '}, 2)
		return lo.T2(parts[0], parts[1])
	})

	byteEncoder := bytesToUnicode()
	byteDecoder := lo.Invert(byteEncoder)

	// translated to a tuple of string, because []byte is not comparable (for maps)
	bpeRank := dictTuple(bpeMerges)

	enc := Encoder{
		bpeRank:     bpeRank,
		encoder:     encoder,
		decoder:     decoder,
		byteEncoder: byteEncoder,
		byteDecoder: byteDecoder,

		cache: sync.Map{},
	}

	return &enc, nil
}

func (e *Encoder) cachedBpe(token string) string {
	cached, ok := e.cache.Load(token)
	if ok {
		return cached.(string)
	}

	output := e.bpe(token)
	e.cache.Store(token, output)
	return output
}

func (e *Encoder) bpe(token string) string {
	word := lo.ChunkString(token, 1)

	pairs := getPairs(word)
	if len(pairs) == 0 {
		return token
	}

	for {
		minPairs := lo.Map(pairs, func(item lo.Tuple2[string, string], index int) lo.Tuple3[int, string, string] {
			pair := lo.T2(item.A, item.B)

			rank, ok := e.bpeRank[pair]
			if !ok {
				rank = math.MaxInt
			}

			return lo.T3(rank, item.A, item.B)
		})

		bigram := lo.MinBy(minPairs, func(a lo.Tuple3[int, string, string], b lo.Tuple3[int, string, string]) bool {
			return a.A < b.A
		})

		first := bigram.B
		second := bigram.C

		if _, ok := e.bpeRank[lo.T2(first, second)]; !ok {
			break
		}

		newWord := []string{}
		for i := 0; i < len(word); {
			_, j, ok := lo.FindIndexOf(word[i:], func(str string) bool { return str == first })
			if !ok {
				newWord = append(newWord, word[i:]...)
				break
			}

			j += i

			newWord = append(newWord, word[i:j]...)
			i = j

			if word[i] == first && i < len(word)-1 && word[i+1] == second {
				newWord = append(newWord, first+second)
				i = i + 2
			} else {
				newWord = append(newWord, word[i])
				i = i + 1
			}
		}

		word = newWord
		if len(word) == 1 {
			break
		}

		pairs = getPairs(word)
	}

	return strings.Join(word, " ")
}

func (e *Encoder) splitToken(token string) ([]string, error) {
	var matches []string

	m, err := pat.FindStringMatch(token)
	if err != nil {
		return nil, err
	}

	for m != nil {
		matches = append(matches, m.String())

		m, err = pat.FindNextMatch(m)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

func (e *Encoder) Encode(text string) ([]int, error) {
	bpeTokens := []int{}

	matches, err := e.splitToken(text)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		runes := []rune(match)

		token := strings.Join(lo.Map(runes, func(item rune, _ int) string {
			return e.byteEncoder[item]
		}), "")

		bpe := e.cachedBpe(token)
		for _, t := range strings.Split(bpe, " ") {
			bpeTokens = append(bpeTokens, e.encoder[t])
		}
	}

	return bpeTokens, nil
}

func (e *Encoder) Decode(tokens []int) string {
	parts := lo.Map(tokens, func(token int, _ int) string {
		return e.decoder[token]
	})

	parts = lo.ChunkString(strings.Join(parts, ""), 1)

	text := lo.Map(parts, func(item string, _ int) rune {
		return e.byteDecoder[item]
	})

	return string(text)
}
