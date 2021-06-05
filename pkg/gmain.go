package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"
)

type writer1 struct {
	writing_buf *[]byte
}

func (w *writer1) write_to_temp_buf(byte byte) {
	*w.writing_buf = append(*w.writing_buf, byte)
	//fmt.Println("WRITER: Written bytes to writing_buf", byte)
}

func (w *writer1) write_to_chan(ch chan []byte) {
	ch <- *w.writing_buf
	//fmt.Println("WRITER: Send byte slice to chan", writing_buf)
	*w.writing_buf = nil
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


func (r *reader1) read_from_chan(ch chan []byte) {
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

func (r *reader1) get20mostfrequentwords() {
	list := make([]int, 20)
	r.rating = &list
	for index, _ := range *r.rating {
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


//I created two structs with the methods to write and read from the shared channel of []bytes. String is not allowed, so we assume that slice of bytes is a word.
//Writer and reader1 works in different goroutines.

//Writer goes through the text and collects bytes to buffer until it reaches the space(which means the end of the word),
//or reaches the symbol which is not a letter(in that case we check do we already have a word in our buffer).
//After this it sends the content of the buffer to channel, and continues to go through the text

//At the same time reader1 listens to the channel. Whenever it gets the word ([]byte) it checks does it have the same word inside the slice of already written words.
//If it does, it increases the counter of this record1, if no, it appends the record1 to it.
//After writer1 and reader1 both finished working with a channel, reader1 goes through the slice of words to determine the most 20 frequent words. I did this with an assumption that,
//any sorting between approximately 74k elements while reading, or quick sorting 10k elements after reading,
//will lead to way more comparisons than going through the slice and just retrieving 20 elements with the largest counter(10k elements in slice, 74k words in text)
//P.S Using one goroutine showed better execution time, idk why, but I decided to go with two goroutines version.
//I used approx 5 hours to code version with no goroutines, and one extra hour to code version that u see

func New(out io.Writer) {
	start := time.Now()
	file, err := ioutil.ReadFile("mobydick.txt") //open file
	if err != nil {
		log.Fatal(err)
	}

	words := make([]record1, 0)
	reader1 := reader1{words: &words} //creating reader1

	writingBuf := make([]byte, 0)
	writer1 := writer1{&writingBuf} //writer1

	ch := make(chan []byte) //channel that we will use to pass slices of bytes from writer1 to reader1
	//btw reader1 listens in range of elements that are passed to channel, it will stop working when there are no elements left, so we don't need any wait groups

	go func() {
		for i := 0; i < len(file)-1; i++ {

			if err == nil {
				byteVal := file[i]
				if byteVal >= 65 && byteVal <= 90 { //if symbol is uppercase letter

					byteVal = byteVal + 32
					writer1.write_to_temp_buf(byteVal) //writing to temporary buffer

				} else if byteVal >= 97 && byteVal <= 122 { //if symbol is lowercase letter

					writer1.write_to_temp_buf(byteVal) //writing to temporary buffer

				} else if byteVal == 32 && len(writingBuf) != 0 { //if symbol is [space], and we have letters in our buffer

					writer1.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer

				} else if ((byteVal > 122 || byteVal < 65) || (byteVal > 90 && byteVal < 97)) && len(writingBuf) != 0 { //if symbol is any other than letter or space, and we have letters in our buffer

					writer1.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer

				} else {
					continue
				}
			} else {
				writer1.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer
				break
			}
		}
		close(ch) //close channel, so our that our reader1 will stop working after there are no elements left, in other case reader1 will cause deadlock
	}()

	reader1.read_from_chan(ch) //reading from channel in range of elements in channel

	reader1.get20mostfrequentwords() //getting 20 most frequent words, and write it to rating slice
	reader1.print()                  //print elements from words according to the rating list
	fmt.Printf("Process took %s\n", time.Since(start))
}