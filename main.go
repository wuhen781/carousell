package main

import (
  "bufio"
  "fmt"
  "os"
  "strings"
  "time"
  "sort"
  "strconv"
  "encoding/csv"
  "errors"
  "github.com/wangjia184/sortedset"
)

//You can set it to 1 if delete_listing happens frequently
const USED_ZSET_FOR_HIGHEST = 0

type User struct {
	data map[string]bool
}

func (u *User) Check(name string) (int,error){
	if _,ok := u.data[name];ok {
		return 0,nil
	}else{
		return 1,errors.New("Error - unknown user")
	}
}

func (u *User) Create(name string) (int,error){
	if _,ok := u.data[name];!ok {
		u.data[name] = true
		return 0,nil
	}else{
		return 1, errors.New("Error - user already existing")
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

func (list *Listing) Delete(user string,id int) (string,error) {
	var item *Item
	var ok bool
	if  item,ok = list.data[id];!ok {
		return "",errors.New("Error - listing does not exist")
	}
	if item.user != user {
		return "",errors.New("Error - listing owner mismatch")
	}
	catname := list.data[id].category
	list.data[id] = nil//to free the memory
	delete(list.data,id);
	return catname,nil
}

func (list *Listing) Get(id int) (*Item,error) {
	var item *Item
	var ok bool
	if  item,ok = list.data[id];!ok {
		return nil , errors.New("Error - not found")
	}
	return item,nil
} 

//format:Phone model 8|Black color, brand new|1000|2019-02-22 12:34:56|Electronics|user1
func (list *Listing) Show(item *Item)  {
	fmt.Printf("%s|%s|%.2f|%s|%s|%s\n",item.title,item.description,item.price,item.createtime.Format("2006-01-02 15:04:05"),item.category,item.user)
}

type Category struct{
	highest_name string
	highest_count int
	data map[string]map[int]*Item
	count map[string]int
	zset *sortedset.SortedSet
}

type CategoryNode struct {
	Name string
	Count int
}

func (cat *Category) AddListing(name string,id int,item *Item)  bool {
	if len(cat.data[name]) < 1{
		cat.data[name] = make(map[int]*Item)
	}
	cat.data[name][id] = item
	cat.count[name] ++
	if cat.count[name] > cat.highest_count {
		cat.highest_count = cat.count[name] 
		cat.highest_name =  name
	}
	if USED_ZSET_FOR_HIGHEST > 0 {
		cat.zset.AddOrUpdate(name, sortedset.SCORE(cat.count[name]), CategoryNode{name,cat.count[name]})//in order to use sortedset,it cost O(logN) to insert
	}
	return true
}

func (cat *Category) DeleteListing(name string,id int)  bool {
	delete(cat.data[name],id);
	cat.count[name] --
	if USED_ZSET_FOR_HIGHEST > 0 {
		//regenerate the highest_count By sortedset O(1)
		if cat.count[name] > 0 {
			cat.zset.AddOrUpdate(name, sortedset.SCORE(cat.count[name]), CategoryNode{name,cat.count[name]})
		}else{
			cat.zset.Remove(name)
		}
		if cat.zset.GetCount() > 0 {
			node:=cat.zset.PeekMax()
			cat.highest_name = node.Value.(CategoryNode).Name
			cat.highest_count =  node.Value.(CategoryNode).Count
		}else{
			cat.highest_name =  ""
			cat.highest_count =  0
		}
	}else{
		//regenerate the highest_count O(n)
		if name == cat.highest_name {
			highest_count := 0
			highest_name := ""
			for k,v := range cat.count {
				if v > highest_count {
					highest_count = v
					highest_name = k
				}
			}
			cat.highest_name = highest_name
			cat.highest_count = highest_count
		}
		//fmt.Println(cat.count)
	}
	return true
}

func (cat *Category) Top() string {
	return cat.highest_name
}

func (cat *Category) Get(name string,sortkey string,order string) ([]*Item,error) {
	arr := make([]*Item,0)
	for _,item := range cat.data[name] {
		arr = append(arr,item)
	}
	if len(arr) > 0 {
		if sortkey == "sort_price" {
			if order == "asc" {
				sort.Slice(arr,func (i,j int) bool {
					return arr[i].price < arr[j].price
				})
			}else{
				sort.Slice(arr,func (i,j int) bool {
					return arr[i].price > arr[j].price
				})
			}
		}else{
			if order == "asc" {
				sort.Slice(arr,func (i,j int) bool {
					return arr[i].createtime.Before(arr[j].createtime)
				})
			}else{
				sort.Slice(arr,func (i,j int) bool {
					return arr[i].createtime.After(arr[j].createtime)
				})
			}
		}
		return arr,nil
	} else {
		return arr , errors.New("Error - category not found")
	}
}

func main() {
	user  := & User{make(map[string]bool)}
	listing := & Listing{100000,make(map[int]*Item)}
	category := & Category{"",0,make(map[string]map[int]*Item),make(map[string]int),sortedset.New()}
	reader := bufio.NewReader(os.Stdin)
	cnt := 0//input count
	for {
		cnt ++
		fmt.Printf("#")
		text,err := reader.ReadString('\n')
		if err != nil {
			return
		}
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "'", "\"", -1)

		// Split string by "
		r := csv.NewReader(strings.NewReader(text))
		r.Comma = ' ' // space
		params, err := r.Read()
		if err != nil {
			fmt.Println(err)
			continue
		}
		operation := params[0]
		if operation == "REGISTER" {
			_,err := user.Create(params[1])
			if err != nil {
				fmt.Println(err)
			}else{
				fmt.Println("Success")
			}
		} else if operation == "CREATE_LISTING" {//input:CREATE_LISTING user1 'Phone model 8' 'Black color, brand new' 1000 'Electronics'
			_,err := user.Check(params[1])
			if err != nil {
				fmt.Println(err)
			}else{
				id := listing.Create(params[1],params[2],params[3],params[4],params[5])
				item,err := listing.Get(id)
				if err != nil {
					fmt.Println(err)
					continue
				}
				category.AddListing(params[5],id,item)
				fmt.Println(id)
			}
		} else if operation == "DELETE_LISTING"  {
			_,err := user.Check(params[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			id, err := strconv.ParseInt(params[2],10,64)
			if err != nil {
				fmt.Println(err)
				continue
			}
			catname,err := listing.Delete(params[1],int(id))
			if err != nil {
				fmt.Println(err)
				continue
			}
			category.DeleteListing(catname,int(id))
			fmt.Println("Success")
		} else if operation == "GET_LISTING"  {
			_,err := user.Check(params[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			id, err := strconv.ParseInt(params[2],10,64)
			if err != nil {
				fmt.Println(err)
				continue
			}
			item , err := listing.Get(int(id))
			if err != nil {
				fmt.Println(err)
				continue
			}
			listing.Show(item)
		} else if operation == "GET_CATEGORY"  {
			_,err := user.Check(params[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			items,err := category.Get(params[2],params[3],params[4])
			if err != nil {
				fmt.Println(err)
				continue
			}
			length := len(items)
			for i:=0;i<length;i++ {
				listing.Show(items[i])
			}
		} else if operation == "GET_TOP_CATEGORY"  {
			_,err := user.Check(params[1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(category.Top())
		} else {
			fmt.Println("unknown operation")
		}
	}
}
