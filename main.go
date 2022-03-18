package main

import (
  "bufio"
  "fmt"
  "os"
  "strings"
  "time"
  "strconv"
)

type User struct {
	data map[string]bool
}

func (u *User) Check(name string) (errno int,errmsg string){
	if _,ok := u.data[name];ok {
		return 0,"Success"
	}else{
		return 1,"Error - unknown user"
	}
}

func (u *User) Create(name string) (errno int,errmsg string){
	if _,ok := u.data[name];!ok {
		u.data[name] = true
		return 0,"Success"
	}else{
		return 1,"Error - user already existing"
	}
}

type Item struct {
	user string
	title string
	description string
	category string
	price float64 
	createtime time.Time
}

type Listing struct {
	primary_id int
	data map[int]*Item
}

func (list *Listing) Create(user,title,description,price,category string) int {
	list.primary_id ++
	id := list.primary_id
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		fmt.Println(err)
	}
	p := & Item {user,title,description,category,priceFloat,time.Now()}
	list.data[id] = p
	return id
}

func (list *Listing) Delete(user string,id int) (errno int,errmsg string) {
return 0,""
}


func main() {
	user  := & User{make(map[string]bool)}
	listing := & Listing{0,make(map[int]*Item)}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("#")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		params := strings.Split(text," ")
		operation := params[0]
		if operation == "REGISTER" {
			errno,errmsg := user.Create(params[1])	
			fmt.Println(errno,errmsg)
		} else if operation == "CREATE_LISTING" {//# CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 'Electronics'
			id := listing.Create(params[1],params[2],params[3],params[4],params[5])	
			fmt.Println(id)
		} else {
	fmt.Println("unknown operation")	
		}
	}
}
