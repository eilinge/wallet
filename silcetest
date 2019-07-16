package main

import "fmt"

// a := copy(s1, s2)
// 		len(s1) > len(s2)
// 		[10 4 5] [1 3 4 5 9]
// 		[1 3 4]

// 		len(s1) < len(s2)
// 		[10 4 5] [1 3]
// 		[1 3 5]

func main() {
	s := [5]int{1, 3, 4, 5, 9}

	s1 := s[:]
	s2 := s[1:4] // [3, 4, 5]

	// copy(s1, s2)
	// len(s1) > len(s2)
	// fmt.Println(s1, s2) // [3 4 5 5 9] [4 5 5]

	copy(s2, s1)
	// len(s2) < len(s1)
	// [3, 4, 5]
	// [1, 3, 4, 5, 9]
	// [1, 3, 4] ->s2=s[1:4] =
	// s -> s[0] + s2 + s[4] = [1 1 3 4 9]
	// copy, append可以改变原来的数组s, s1, s2基于已经改变了s=[1 1 3 4 9]进行分割
	// s1 = s[:]
	// s2 = s[1:4]
	// copy 切片改变了原来的数组
	fmt.Println(s1, s2) // [1 1 3 4 9] [1 3 4]

}
