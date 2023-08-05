package main

import (
	"github.com/manishchauhan/dugguGo/util"
	"github.com/manishchauhan/dugguGo/util/tree"
)

func resolveTree() {

	t := tree.Tree{}

	t.Insert(50)
	t.Insert(30)
	t.Insert(20)
	t.Insert(40)
	t.Insert(70)
	t.Insert(60)
	t.Insert(80)

	t.InOrderTraversal()
}

func main() {

	//result := util.Addnumber(40, 90)
	//fmt.Println(result)
	//resolveTree()
	var manish util.User
	manish.Age = 22
	println(manish.Age)
	println(util.MakeSumArray())
}
