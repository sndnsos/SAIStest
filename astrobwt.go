package  astrobwt

//import "fmt"
import "errors"
import "sort"
import "github.com/shenwei356/util/byteutil"

//import "math/rand"
import "golang.org/x/crypto/sha3"
import "encoding/binary"

import "golang.org/x/crypto/salsa20/salsa"

// see here to improve the algorithms more https://github.com/y-256/libdivsufsort/blob/wiki/SACA_Benchmarks.md

// ErrInvalidSuffixArray means length of sa is not equal to 1+len(s)
var ErrInvalidSuffixArray = errors.New("bwt: invalid suffix array")

// Transform returns Burrows–Wheeler transform of a byte slice.
// See https://en.wikipedia.org/wiki/Burrows%E2%80%93Wheeler_transform
func Transform(s []byte, es byte) ([]byte, error) {
	sa := SuffixArray(s)
	bwt, err := FromSuffixArray(s, sa, es)
	return bwt, err
}

// InverseTransform reverses the bwt to original byte slice. Not optimized yet.
func InverseTransform(t []byte, es byte) []byte {
	n := len(t)
	lines := make([][]byte, n)
	for i := 0; i < n; i++ {
		lines[i] = make([]byte, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			lines[j][n-1-i] = t[j]
		}
		sort.Sort(byteutil.SliceOfByteSlice(lines))
	}

	s := make([]byte, n-1)
	for _, line := range lines {
		if line[n-1] == es {
			s = line[0 : n-1]
			break
		}
	}
	return s
}

// SuffixArray returns the suffix array of s.
func SuffixArray(s []byte) []int {
	_sa := New(s)
	var sa []int = make([]int, len(s)+1)
	sa[0] = len(s)
	for i := 0; i < len(s); i++ {
		sa[i+1] = int(_sa.sa.int32[i])
	}
	return sa
}

// FromSuffixArray compute BWT from sa
func FromSuffixArray(s []byte, sa []int, es byte) ([]byte, error) {
	if len(s)+1 != len(sa) || sa[0] != len(s) {
		return nil, ErrInvalidSuffixArray
	}
	bwt := make([]byte, len(sa))
	bwt[0] = s[len(s)-1]
	for i := 1; i < len(sa); i++ {
		if sa[i] == 0 {
			bwt[i] = es
		} else {
			bwt[i] = s[sa[i]-1]
		}
	}
	return bwt, nil
}

func BWT(input []byte) ([]byte, int) {
	if len(input) >= maxData32 {
		panic("input too big to handle")
	}
	sa := make([]int32, len(input)+1)
	text_32(input, sa[1:])

	bwt := make([]byte, len(input)+1)
	bwt[0] = input[len(input)-1]
	emarker := 0
	for i := 1; i < len(sa); i++ {
		if sa[i] == 0 {
			//bwt[i] = '$' //es
			emarker = i
		} else {
			bwt[i] = input[sa[i]-1]
		}

	}
	//bwt[emarker] = '$'
	return bwt, emarker
}

const stage1_length int = 69371 // it is a prime
const max_length int = 1024*1024 + stage1_length + 1024

func POW(inputdata []byte) (outputhash [32]byte) {

	var counter [16]byte

	key := sha3.Sum256(inputdata)

	var stage1 [stage1_length]byte                    // stages are taken from it
	var stage2 [1024*1024 + stage1_length + 1024]byte // it will always be less than 1 MB  0x7fff * 31 + stage1_length

	salsa.XORKeyStream(stage1[:stage1_length], stage1[:stage1_length], &counter, &key)

	stage1_result, eos := BWT(stage1[:stage1_length])

	key = sha3.Sum256(stage1_result)

	stage2_length := stage1_length + int(binary.LittleEndian.Uint32(key[:])&0xfffff)

	for i := range counter { // will be optimized by compiler
		counter[i] = 0
	}

	salsa.XORKeyStream(stage2[:stage2_length], stage2[:stage2_length], &counter, &key)

	stage2_result, eos := BWT(stage2[:stage2_length])

	//	fmt.Printf("result %x  stage2_length %d \n", key, stage2_length)
	key = sha3.Sum256(stage2_result)

	//fmt.Printf("result %x\n", key)

	_ = eos
	return
}

// input byte
// sa should be len(input) + 1
// result len len(input) + 1
func BWT_0alloc(input []byte, sa []int32, bwt []byte) int {

	//ix := &Index{data: input}
	if len(input) >= maxData32 {
		panic("input too big to handle")
	}
	if len(sa) != len(input)+1 {
		panic("invalid sa array")
	}
	if len(bwt) != len(input)+1 {
		panic("invalid bwt array")
	}
	//sa := make([]int32, len(input)+1)
	text_32(input, sa[1:])

	//bwt := make([]byte, len(input)+1)
	bwt[0] = input[len(input)-1]
	emarker := 0
	for i := 1; i < len(sa); i++ {
		if sa[i] == 0 {
			//bwt[i] = '$' //es
			emarker = i
		} else {
			bwt[i] = input[sa[i]-1]
		}

	}
	//bwt[emarker] = '$'
	return emarker
}
func POW_0alloc(inputdata []byte) (outputhash [32]byte) {

	var counter [16]byte

	var sa [max_length]int32
	// var bwt [max_length]int32
	var stage2 [1024*1024 + stage1_length + 1024]byte // it will always be less than 1 MB  0x7fff * 31 + stage1_length

	key := sha3.Sum256(inputdata)

	//var stage1 [stage1_length]byte // stages are taken from it
	//var stage1_result [stage1_length + 1]byte
	//salsa.XORKeyStream(stage1[:stage1_length], stage1[:stage1_length], &counter, &key)
	//eos := BWT_0alloc(stage1[:stage1_length], sa[:stage1_length+1], stage1_result[:stage1_length+1])
	//key = sha3.Sum256(stage1_result[:])

	stage2_length := stage1_length + int(binary.LittleEndian.Uint32(key[:])&0xfffff)
	//println(stage2_length)
	for i := range counter { // will be optimized by compiler
		counter[i] = 0
	}
	salsa.XORKeyStream(stage2[:stage2_length], stage2[:stage2_length], &counter, &key)
	for i := range sa {
		sa[i] = 0
	}

	var stage2_result [1024*1024 + stage1_length + 1024 + 1]byte
	eos := BWT_0alloc(stage2[:stage2_length], sa[:stage2_length+1], stage2_result[:stage2_length+1])
	// key = sha3.Sum256(stage2_result[:])

	_ = eos
	return
}
