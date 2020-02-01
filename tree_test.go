package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"testing"
)

func getPhone(ctx *Context){
	fmt.Println("get Phone")
}

func getPath(ctx *Context){
	fmt.Println("get path")
}

func getSex(ctx *Context){
	fmt.Println("get Sex")
}

func getSexaby(ctx *Context){
	fmt.Println("get Sexaby")
}

func getSexayy(ctx *Context){
	fmt.Println("get Sexayy")
}

func Print(rootNode *Node){
	fmt.Println(rootNode.matchChars)
	if len(rootNode.childNodes) > 0{
		for _,node := range rootNode.childNodes{
			Print(node)
			if node.nodeType == NodeTypeParam || node.nodeType == NodeTypeAll{
				fmt.Println("wild",node.wildcardParams)
			}
		}
	}
}

func TestTree(t *testing.T){
	rootNode := &Node{nodeType:NodeTypeRoot,handler:nil,wildcardParams:make([]string,0),childNodes:make([]*Node,0)}
	rootNode.Add("/getPhone/:name",ContextHandler(getPhone))
	rootNode.Add("/getSex",ContextHandler(getSex))
	rootNode.Add("/getSexaby",ContextHandler(getSexaby))
	rootNode.Add("/getSexayy",ContextHandler(getSexayy))
	Print(rootNode)
	h,_:=rootNode.Find("/getPhone/yilin",func(k string,v interface{}){
		fmt.Println(k,v)
	})
	if nil != h{
			h(nil)
	}
}