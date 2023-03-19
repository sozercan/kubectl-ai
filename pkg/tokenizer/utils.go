package tokenizer

import (
	"github.com/samber/lo"
)

func dictTuple(tuples []lo.Tuple2[[]byte, []byte]) map[lo.Tuple2[string, string]]int {
	i := -1
	return lo.SliceToMap(tuples, func(item lo.Tuple2[[]byte, []byte]) (lo.Tuple2[string, string], int) {
		i++
		return lo.T2(string(item.A), string(item.B)), i
	})
}
