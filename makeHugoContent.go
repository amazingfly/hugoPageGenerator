package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/TeamFairmont/boltsdk-go/boltsdk"
	"github.com/TeamFairmont/boltshared/mqwrapper"
	"github.com/TeamFairmont/gabs"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func startWork() {
	var mq *mqwrapper.Connection
	var err error
	connected := false
	attempts := 0

	for !connected {

		//connect to mq server
		mq, err = mqwrapper.ConnectMQ("amqp://guest:guest@192.168.56.105:5672/")
		if err != nil {
			if attempts > 5 {

				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			connected = true
		}
		attempts++
		time.Sleep(1000000)
	}

	makeHugoCMD := "/createHugoFiles"

	//set the name of our command
	boltsdk.EnableLogOutput(true)

	boltsdk.RunWorker(mq, "", makeHugoCMD, MakeHugoWork)

	forever := make(chan bool)
	<-forever
}

// MakeHugoWork imports json and creates the hugo content file
func MakeHugoWork(payload *gabs.Container) error {
	var err error
	fmt.Println("yodizzle")
	return err
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

const numberBytes = "12345" //"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	numberIdxBits = 6                    // 6 bits to represent a letter index
	numberIdxMask = 1<<numberIdxBits - 1 // All 1-bits, as many as letterIdxBits
	numberIdxMax  = 63 / numberIdxBits   // # of letter indices fitting in 63 bits
)

func RandomInt(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), numberIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), numberIdxMax
		}
		if idx := int(cache & numberIdxMask); idx < len(numberBytes) {
			b[i] = numberBytes[idx]
			i--
		}
		cache >>= numberIdxBits
		remain--
	}

	return string(b)
}

func main() {

	for i := 0; i < 100; i++ {
		makeHugoFiles()
	}
	forever := make(chan bool)
	<-forever
}
func makeHugoFiles() {

	//var hugoVariables string
	date := `"` + RandomInt(4) + "-0" + RandomInt(1) + "-0" + RandomInt(1) + "T0" + RandomInt(1) + ":" + RandomInt(2) + ":" + RandomInt(2) + "-" + RandomInt(2) + ":" + RandomInt(2) + `"`
	title := RandomString(10) + "za"
	draft := "false"
	image := `"` + title + ".jpg" + `"`
	price := RandomInt(2) + "." + RandomInt(2)
	sku := `"` + RandomString(2) + RandomInt(3) + `"`
	hugoVariables := []string{"date = " + date, `title = "` + title + `"`, "draft = " + draft, "image = " + image, "price = " + price, "sku = " + sku}
	// will store the contents of the hugo content files content
	var hugoContent string
	hugoContent = RandomString(250)

	go makeCatPicture(title)

	//open new file
	f, err := os.Create("../bookshelf/content/post/" + title + ".md")
	check(err)
	defer f.Close()

	//creat buffered writer to write to file
	w := bufio.NewWriter(f)
	//write "+++\n" to top of the hugo content file
	_, err = w.Write([]byte{43, 43, 43, 10}) // +++\n
	check(err)

	for _, b := range hugoVariables {
		//write the post variables
		_, err = w.WriteString(b)
		_, err = w.Write([]byte{10})
		check(err)
	}
	//write "+++\n" below the variables of the hugo content file
	_, err = w.Write([]byte{43, 43, 43, 10}) // +++\n
	check(err)

	//write the post content
	_, err = w.WriteString(hugoContent)
	check(err)

	w.Flush()
}

func makeCatPicture(title string) {
	cmd := exec.Command("wget", "-O", "../bookshelf/static/images/"+title+".jpg", "http://placekitten.com/"+RandomInt(2)+"/"+RandomInt(2))
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
}
