package util

//import "sort"

const LOCALHOST = "127.0.0.1"

func Copy(data int, i int) int {
	return data
}

// No generics or polymorphism
func CopyString(data string, i int) string {
	return data
}

func CopyBates(data []byte, i int) []byte {
	bs := make([]byte, len(data))
	copy(bs, data)
	return bs
}

func Sum(xs []int) int {
	res := 0
	for i := 0; i < len(xs); i++ {
		res = res + xs[i]
	}
	return res
}

func UnaryReduce(xs []int) int {
	return xs[0]
}

func UnaryReduceString(xs []string) string {
	return xs[0]
}

func UnaryReduceBates(xs [][]byte) []byte {
	return xs[0]
}

/*//func GetValues(m map[int] int) []int {
func GetValues(m map[int] interface{}) []interface{} {
	xs := make([]interface{}, len(m))
	keys := make([]int, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for i, k := range keys {
		xs[i] = m[k]
	}
	return xs
}

func GetValuesInt(m map[int] int) []int {
	xs := make([]int, len(m))
	keys := make([]int, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for i, k := range keys {
		xs[i] = m[k]
	}
	return xs
}

func GetValuesString(m map[int] string) []string {
	xs := make([]string, len(m))
	keys := make([]int, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for i, k := range keys {
		xs[i] = m[k]
	}
	return xs
}

func GetValuesBates(m map[int] []byte) [][]byte {
	xs := make([][]byte, len(m))
	keys := make([]int, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for i, k := range keys {
		xs[i] = m[k]
	}
	return xs
}*/
