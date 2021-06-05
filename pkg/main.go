package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type writer struct {
	writing_buf *[]byte
}

func (w *writer) write_to_temp_buf(byte byte) {
	*w.writing_buf = append(*w.writing_buf, byte)
	//fmt.Println("WRITER: Written bytes to writing_buf", byte)
}

func (w *writer) write_to_chan(ch chan []byte) {
	ch <- *w.writing_buf
	//fmt.Println("WRITER: Send byte slice to chan", writing_buf)
	*w.writing_buf = nil
}

type reader struct {
	words  *[]record
	rating *[]int
}

type record struct {
	word    []byte
	counter int
	checked bool
}

func (r *reader) contains(element []byte) (bool, int) {
	for index, v := range *r.words {
		if bytes.Equal(v.word, element) {
			return true, index
		}
		index = index + 1
	}
	return false, 0
}

func (r *reader) read_from_chan(ch chan []byte) {
	for node := range ch {
		state, index := r.contains(node)
		if state {
			(*r.words)[index].counter++
		} else {
			record := record{node, 1, false}
			*r.words = append(*r.words, record)
		}
	}
}

func (r *reader) get20mostfrequentwords() {
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

func (r *reader) print() {
	for _, v := range *r.rating {
		fmt.Print((*r.words)[v].counter, " ", string((*r.words)[v].word))
		fmt.Println()

	}
}

//I created two structs with the methods to write and read from the shared channel of []bytes. String is not allowed, so we assume that slice of bytes is a word.
//Writer and reader works in different goroutines.

//Writer goes through the text and collects bytes to buffer until it reaches the space(which means the end of the word),
//or reaches the symbol which is not a letter(in that case we check do we already have a word in our buffer).
//After this it sends the content of the buffer to channel, and continues to go through the text

//At the same time reader listens to the channel. Whenever it gets the word ([]byte) it checks does it have the same word inside the slice of already written words.
//If it does, it increases the counter of this record, if no, it appends the record to it.
//After writer and reader both finished working with a channel, reader goes through the slice of words to determine the most 20 frequent words. I did this with an assumption that,
//any sorting between approximately 74k elements while reading, or quick sorting 10k elements after reading,
//will lead to way more comparisons than going through the slice and just retrieving 20 elements with the largest counter(10k elements in slice, 74k words in text)
//P.S Using one goroutine showed better execution time, idk why, but I decided to go with two goroutines version.
//I used approx 5 hours to code version with no goroutines, and one extra hour to code version that u see

func Old(out io.Writer) {

	start := time.Now()
	file, err := os.Open("mobydick.txt") //open file
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	readingBuf := make([]byte, 1) //read file by one letter only

	words := make([]record, 0)
	reader := reader{words: &words} //creating reader

	writingBuf := make([]byte, 0)
	writer := writer{&writingBuf} //writer

	ch := make(chan []byte) //channel that we will use to pass slices of bytes from writer to reader
	//btw reader listens in range of elements that are passed to channel, it will stop working when there are no elements left, so we don't need any wait groups

	go func() {
		for {
			//reading file's letters one by one
			n, err := file.Read(readingBuf)

			if n > 0 {
				byteVal := readingBuf[0]
				if byteVal >= 65 && byteVal <= 90 { //if symbol is uppercase letter

					byteVal = byteVal + 32
					writer.write_to_temp_buf(byteVal) //writing to temporary buffer

				} else if byteVal >= 97 && byteVal <= 122 { //if symbol is lowercase letter

					writer.write_to_temp_buf(byteVal) //writing to temporary buffer

				} else if byteVal == 32 && len(writingBuf) != 0 { //if symbol is [space], and we have letters in our buffer

					writer.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer

				} else if ((byteVal > 122 || byteVal < 65) || (byteVal > 90 && byteVal < 97)) && len(writingBuf) != 0 { //if symbol is any other than letter or space, and we have letters in our buffer

					writer.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer

				} else {
					continue
				}
			}

			if err == io.EOF {
				writer.write_to_chan(ch) //send temporary buffer content to channel, empty the temporary buffer
				break
			}
		}
		close(ch) //close channel, so our that our reader will stop working after there are no elements left, in other case reader will cause deadlock
	}()

	reader.read_from_chan(ch) //reading from channel in range of elements in channel

	reader.get20mostfrequentwords() //getting 20 most frequent words, and write it to rating slice
	reader.print()                  //print elements from words according to the rating list
	fmt.Printf("Process took %s\n", time.Since(start))
}
