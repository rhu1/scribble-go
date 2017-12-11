package util


const LOCALHOST = "127.0.0.1"


func Copy(data int, i int) int {
	return data
}

func Sum(xs []int) int {
	res := 0
	for i := 0; i < len(xs); i++ {
		res = res + xs[i]	
	}
	return res
}
