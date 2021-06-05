package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

type writer1 struct {
	writingBuf *[]byte
}

func (w *writer1) writeToTempBuf(byte byte) {
	*w.writingBuf = append(*w.writingBuf, byte)
}

func (w *writer1) writeToChan(ch chan []byte) {
	ch <- *w.writingBuf
	*w.writingBuf = nil
}

type reader1 struct {
	words  *[]record1
	rating *[]int
}

type record1 struct {
	word    []byte
	counter int
	checked bool
}

func (r *reader1) contains(element []byte) (bool, int) {
	for index, v := range *r.words {
		if bytes.Compare(v.word, element) == 0 {
			return true, index
		}
		index = index + 1
	}
	return false, 0
}

func (r *reader1) readFromChan(ch chan []byte) {
	for node := range ch {
		state, index := r.contains(node)
		if state {
			(*r.words)[index].counter++
		} else {
			record1 := record1{node, 1, false}
			*r.words = append(*r.words, record1)
		}
	}
}

func (r *reader1) getMostFrequentWords(number int) {
	list := make([]int, number)
	r.rating = &list
	for index := range *r.rating {
		temp := 0
		inss := 0
		for index, v := range *r.words {
			if (v.checked == false) && (v.counter > temp) {
				temp = v.counter
				inss = index
			}
		}
		(*r.words)[inss].checked = true
		(*r.rating)[index] = inss
	}
}

func (r *reader1) print() {
	for _, v := range *r.rating {
		fmt.Print((*r.words)[v].counter, " ", string((*r.words)[v].word))
		fmt.Println()
	}
}

func main() {
	start := time.Now()
	file, err := ioutil.ReadFile("pkg/mobydick.txt")
	if err != nil {
		log.Fatal(err)
	}

	words := make([]record1, 0)
	reader1 := reader1{words: &words}

	writingBuf := make([]byte, 0)
	writer1 := writer1{&writingBuf}

	ch := make(chan []byte)

	go func() {
		for i := 0; i < len(file)-1; i++ {
			if err == nil {
				byteVal := file[i]
				switch {
				case byteVal >= 65 && byteVal <= 90:
					writer1.writeToTempBuf(byteVal + 32)
				case byteVal >= 97 && byteVal <= 122:
					writer1.writeToTempBuf(byteVal)
				case byteVal > 0:
					writer1.writeToChan(ch)
				default:
					continue
				}
			} else {
				writer1.writeToChan(ch)
				break
			}
		}
		close(ch)
	}()

	reader1.readFromChan(ch)

	reader1.getMostFrequentWords(20)
	reader1.print()
	fmt.Printf("Process took %s\n", time.Since(start))
}
