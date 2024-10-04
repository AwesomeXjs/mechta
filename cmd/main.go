package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

/*

Обьяснение решения:
Реализовал пул воркеров на основе буферизированного канала.
Буфер канала равен максимальному количеству горутин,
когда он забитый - оставшиеся горутины ждут своей очереди.

NewDecoder - декодирует JSON по частям и будет более эффективный на большом обьеме данных.

Для атомарности использовал Atomic, так как мы работаем с числами.

*/

const (
	filePath      = "nums.json"
	maxGoroutines = 10
)

type Element struct {
	A int `json:"a"`
	B int `json:"b"`
}

func main() {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(file)
	var elements []Element
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&elements); err != nil {
		log.Fatal(err.Error())
	}

	sum, err := sumNums(elements, maxGoroutines)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(sum)
}

// sumNums - суммирует числа в массиве, запуская отдельную горутину для каждого элемента
func sumNums(elements []Element, maxGoroutines int) (int64, error) {
	var sum int64
	var wg sync.WaitGroup
	myChan := make(chan struct{}, maxGoroutines)

	for _, element := range elements {
		wg.Add(1)
		if element.A < -10 || element.A > 10 || element.B < -10 || element.B > 10 {
			wg.Done()
			return 0, fmt.Errorf("неверный массив, нужны числа от -10 до 10")
		}
		go func(el Element) {
			defer wg.Done()
			worker(el, myChan, &sum)
		}(element)
	}
	wg.Wait()
	return sum, nil
}

// worker - выполняет сложение двух чисел атомарно при помощи atomic.AddInt64
func worker(el Element, myChan chan struct{}, sum *int64) {
	myChan <- struct{}{}

	atomic.AddInt64(sum, int64(el.A+el.B))

	<-myChan

}
