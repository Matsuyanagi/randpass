package main

import (
	cryptRand "crypto/rand" //	乱数 : math/rand と名前がかぶるためにエイリアスする
	"encoding/binary"
	"flag" //	コマンドライン引数解析
	"fmt"
	"io"        //	文字出力
	"math/rand" //	乱数
	"os"
)

// 含まれる文字種
const (
	bitAlphaCapital = 1 << iota // 大文字アルファベット
	bitAlphaSmall               // 小文字アルファベット
	bitAlphaNum                 // 数字
	bitAlphaSymbol              // 記号
)

// デフォルト数量
const (
	defaultPassStringLength  = 12 // デフォルト文字列長
	defaultPassStringNumbers = 1  // デフォルト文字列数
)

// 使用文字種。視認性の低い文字をのぞいておく
//	runeArrayAlphaCapital := []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var runeArrayAlphaCapital = []rune{'A', 'B', 'C', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'R', 'T', 'U', 'V', 'W', 'X', 'Y'}

//	runeArrayAlphaSmall := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
var runeArrayAlphaSmall = []rune{'a', 'b', 'c', 'd', 'e', 'f', 'h', 'i', 'j', 'k', 'm', 'n', 'p', 'r', 't', 'u', 'w', 'y'}

//	runeArrayNum := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var runeArrayNum = []rune{'2', '3', '4', '5', '6', '7', '8', '9'}

//	runeArraySymbol := []rune{'!', '#', '$', '%', '&', '(', ')', '-', '=', '^', '~', '<', '>', '@', '[', ']', ';', ':', '?'}
var runeArraySymbol = []rune{'!', '#', '$', '%', '&', '(', ')', '-', '=', '^', '<', '>', '@', '[', ']', ':', '?'}

func main() {
	// 乱数初期化
	InitRandom()

	// コマンドライン引数パース
	var passStringLength int
	var passStringNumbers int
	var flagKindCS bool
	var flagKindCSN bool
	var flagKindCSNY bool
	flag.IntVar(&passStringLength, "L", defaultPassStringLength, "文字列長")
	flag.IntVar(&passStringNumbers, "N", defaultPassStringNumbers, "文字列数")
	flag.BoolVar(&flagKindCS, "S", false, "文字種[アルファベット大文字,小文字]")
	flag.BoolVar(&flagKindCSN, "C", false, "文字種[アルファベット大文字,小文字,数字]")
	flag.BoolVar(&flagKindCSNY, "Y", false, "文字種[アルファベット大文字,小文字,数字,記号]")
	flag.Parse()

	//	何も指定されてなければ flagKindCSN が指定されたことにする
	if !(flagKindCS || flagKindCSN || flagKindCSNY) {
		flagKindCSN = true
	}

	flagStringKinds := 0
	if flagKindCS {
		flagStringKinds |= bitAlphaCapital | bitAlphaSmall
	}
	if flagKindCSN {
		flagStringKinds |= bitAlphaCapital | bitAlphaSmall | bitAlphaNum
	}
	if flagKindCSNY {
		flagStringKinds |= bitAlphaCapital | bitAlphaSmall | bitAlphaNum | bitAlphaSymbol
	}

	// 指定出力先へ文字種、長さ、個数を指定して出力する
	CreatePassString(os.Stdout, flagStringKinds, passStringLength, passStringNumbers)

}

// CreatePassString 指定出力先へ文字種、長さ、個数を指定して出力する
func CreatePassString(w io.Writer, flagStringKinds int, passStringLength int, passStringNumbers int) {
	// srcRuneArrayへの配列の追加は1回だけにしたい
	flagFirstTime := true

	// 文字列種類を問わずに最後に抽選する文字列種
	srcRuneArray := make([]rune, 0, 100)

	for i := 0; i < passStringNumbers; i++ {

		stringLength := passStringLength

		selectedRunes := make([]rune, 0, 100)
		if flagStringKinds&bitAlphaCapital != 0 {
			// 大文字アルファベット1文字抽選
			selectedRunes = append(selectedRunes, SelectRandomArrayRune(runeArrayAlphaCapital))
			if flagFirstTime {
				srcRuneArray = append(srcRuneArray, runeArrayAlphaCapital...)
			}
			stringLength--
		}
		if flagStringKinds&bitAlphaSmall != 0 {
			// 小文字アルファベット1文字抽選
			selectedRunes = append(selectedRunes, SelectRandomArrayRune(runeArrayAlphaSmall))
			if flagFirstTime {
				srcRuneArray = append(srcRuneArray, runeArrayAlphaSmall...)
			}
			stringLength--
		}
		if flagStringKinds&bitAlphaNum != 0 {
			// 数字1文字抽選
			selectedRunes = append(selectedRunes, SelectRandomArrayRune(runeArrayNum))
			if flagFirstTime {
				srcRuneArray = append(srcRuneArray, runeArrayNum...)
			}
			stringLength--
		}
		if flagStringKinds&bitAlphaSymbol != 0 {
			// 記号1文字抽選
			selectedRunes = append(selectedRunes, SelectRandomArrayRune(runeArraySymbol))
			if flagFirstTime {
				srcRuneArray = append(srcRuneArray, runeArraySymbol...)
			}
			stringLength--
		}
		flagFirstTime = false

		//	残りの文字を入れる
		for ; stringLength > 0; stringLength-- {
			selectedRunes = append(selectedRunes, SelectRandomArrayRune(srcRuneArray))
		}

		// 文字配列から文字列に変換。文字種の方が文字列長より長い場合も考慮して末尾を切る
		resultString := string(ShuffleRuneArray(selectedRunes)[:passStringLength])

		// 指定出力先へ出力
		fmt.Fprintln(w, resultString)
	}
}

// InitRandom 乱数を初期化する
// https://golang.org/pkg/math/rand/#Seed
// 2 ^ 31-1で除算したときに同じ剰余を持つシード値は、同じ擬似ランダムシーケンスを生成します
//	rand.Seed(2147483647)	( 2 ** 31 - 1 ) * 1
//	rand.Seed(4294967294)	( 2 ** 31 - 1 ) * 2
//	rand.Seed(6442450941)	( 2 ** 31 - 1 ) * 3
//		→これらは同じ乱数系列になる
func InitRandom() {
	var n int64
	binary.Read(cryptRand.Reader, binary.LittleEndian, &n)
	//	math.rand.Seed(time.Now().UnixNano())
	rand.Seed(n)
}

// SelectRandomArrayRune rune 配列からランダムで1文字を返す
func SelectRandomArrayRune(array []rune) rune {
	length := len(array)
	r := rand.Intn(length)
	return array[r]
}

// ShuffleRuneArray rune 配列をシャッフルする
func ShuffleRuneArray(array []rune) []rune {
	for i := len(array); i > 1; i-- {
		j := rand.Intn(i) // 0～(i-1) の乱数発生
		array[i-1], array[j] = array[j], array[i-1]
	}
	return array
}
