package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TeamFairmont/boltsdk-go/boltsdk"
	"github.com/TeamFairmont/boltshared/mqwrapper"
	"github.com/TeamFairmont/gabs"
)

var (
	recordChan = make(chan map[string]string, 2)
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
	readProducts()
	fmt.Println("done it seems")
	time.Sleep(2 * time.Second)
	fmt.Println("done it seems")
	forever := make(chan bool)
	<-forever

}

/*
SKU
RetailPrice1
SpecialPrice1
ProdName
ProdInentoory
ProdDescription
ProductURLName
ProdStatus
Unit
PackageHeight
PackageLength
PackageWidth
ActualWeight
MinimumQuantity
*/
func makeHugoFiles(c map[string]interface{}) {
	//record := <-recordChan
	record := c
	var title, price, sku, image, hugoContent string
	//var hugoVariables string
	date := `"` + RandomInt(4) + "-0" + RandomInt(1) + "-0" + RandomInt(1) + "T0" + RandomInt(1) + ":" + RandomInt(2) + ":" + RandomInt(2) + "-" + RandomInt(2) + ":" + RandomInt(2) + `"`
	if fieldExists(record, "ProdName") {
		if strings.Contains(record["ProdName"].(string), ` `) {
			record["ProdName"] = strings.Replace(record["ProdName"].(string), ` `, `_`, -1)
		}
		if strings.Contains(record["ProdName"].(string), `"`) {
			record["ProdName"] = strings.Replace(record["ProdName"].(string), `"`, ``, -1)
		}
		if strings.Contains(record["ProdName"].(string), `.`) {
			record["ProdName"] = strings.Replace(record["ProdName"].(string), `.`, ``, -1)
		}
		if strings.Contains(record["ProdName"].(string), `%`) {
			record["ProdName"] = strings.Replace(record["ProdName"].(string), `%`, ``, -1)
		}
		title = record["ProdName"].(string)
	}
	draft := "false"

	image = `"` + title + ".jpg" + `"`

	if fieldExists(record, "RetailPrice1") {
		price = record["RetailPrice1"].(string)
	}
	if fieldExists(record, "SKU") {
		sku = `"` + record["SKU"].(string) + `"`
	}
	hugoVariables := []string{"date = " + date, `title = "` + title + `"`, "draft = " + draft, "image = " + image, "price = " + price, "sku = " + sku}
	// will store the contents of the hugo content files content
	if fieldExists(record, "ProdDescription") {
		hugoContent = record["ProdDescription"].(string)
	}

	go makeCatPicture(title)
	fmt.Println("opening file: ", title)
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

// reads the csv file and creates a array of maps, each array item represents a record / page / product
func readProducts() {
	var catCount = 0
	//open csv file
	file, err := os.Open(`cv3_product_export1.csv`)
	if err != nil {
		fmt.Println(err)
	}
	//read file
	r := bufio.NewReader(file)
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 //set to unlimited
	//Read just the first section, which is the fields of the csv spreadsheet
	fields, err := reader.Read()
	if err != nil {
		fmt.Println(err)
	}
	//read the rest of the file, which is product data
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	//start building a JSON array of objects
	var buffer bytes.Buffer
	buffer.WriteString(`[`)
	for _, rec := range records { //range over each record
		buffer.WriteString(`{`)
		for y, rc := range rec { //rage over each fild in a record
			buffer.WriteString(`"`)
			buffer.WriteString(fields[y]) //add key name
			buffer.WriteString(`": "`)
			//if any quotes need to be escaped
			if strings.Contains(rc, `"`) {
				rc = strings.Replace(rc, `"`, `\"`, -1)
			} //if any new lines need removed
			if strings.Contains(rc, "\n") {
				rc = strings.Replace(rc, "\n", "", -1)
			}
			if strings.Contains(rc, `<sup>`) {
				rc = strings.Replace(rc, `<sup>`, "", -1)
			}
			if strings.Contains(rc, `&reg;`) {
				rc = strings.Replace(rc, `&reg;`, "", -1)
			}
			if strings.Contains(rc, `</sup>`) {
				rc = strings.Replace(rc, `</sup>`, "", -1)
			}
			if strings.Contains(rc, `(`) {
				rc = strings.Replace(rc, `(`, "", -1)
			}
			if strings.Contains(rc, `)`) {
				rc = strings.Replace(rc, `)`, "", -1)
			}
			if strings.Contains(rc, `/`) {
				rc = strings.Replace(rc, `/`, `-`, -1)
			}
			buffer.WriteString(rc) //add data
			buffer.WriteString(`",`)
		} //finished ranging over fields of a record
		buffer.Truncate(len(buffer.Bytes()) - 1) // remove last camma ","
		buffer.WriteString(`},`)
	} // finished with records
	buffer.Truncate(len(buffer.Bytes()) - 1) // remove last camma ","
	buffer.WriteString(`]`)

	//make empty interface
	var f interface{}
	err = json.Unmarshal(buffer.Bytes(), &f) //fill interface with json
	if err != nil {
		fmt.Println(err)
	}
	var m []interface{}
	if f != nil {
		m = f.([]interface{}) //type assert the interface into an array of interfaces
	}
	count := 0 //count total records
	//var list = make(map[string]int) // TODO was used for field inspection, maybe not needed
	//range over array of interfaces
	for _, b := range m {
		c := b.(map[string]interface{}) //assert that b of type interface{} is of type map;string'interface{}
		//recordChan <- c
		//send record to makeHugoContent()

		catCount++
		if catCount%100 == 0 {
			fmt.Println("hold on! There is a cat jam!")
			time.Sleep(2 * time.Second)
		}
		go makeHugoFiles(c)
		count++
		/*
			        //TODO mightnot be used anymore, was for examining the record data
					for q, w := range c {
						if w.(string) != "" {
							//checkSlice(list, q)
						}
					}
		*/
	}
	//fmt.Println(xc)
	//sortMap(list)

	fmt.Println(count)
}

//TODO was used for examining record data
func checkSlice(list map[string]int, key string) {
	_, ok := list[key]
	if !ok {
		list[key] = 1
		fmt.Println(key + " added")
	}

	for a := range list {
		if a == key {
			list[key]++
		}
	}
}

//TODO was used for examining record data
func sortMap(list map[string]int) {
	count := 0
	for x, y := range list {

		_, ok := list[x]
		if ok && y > 1320 {
			count++
			fmt.Println(fmt.Sprint(y) + " : " + x)
		}
	}
	fmt.Println(count)
	/*
		    var order = 3333
			var reverse = make(map[int]string)

			for x, y := range list {
				reverse[y] = x
				//fmt.Println("added " + fmt.Sprint(y))
			}

			for order >= 0 {
				//fmt.Println("loopin"+fmt.Sprint(order))
				_, ok := reverse[order]
				if ok {
					//fmt.Println(ok)
					fmt.Println(fmt.Sprint(order) + " : " + reverse[order])
				}
				order--
			}
	*/
}
func fieldExists(m map[string]interface{}, s string) bool {
	_, ok := m[s]
	return ok
}
