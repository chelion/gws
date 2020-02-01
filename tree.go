package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"errors"
	"github.com/chelion/gws/utils"
)

type NodeType uint8

const(
	NodeTypeRoot  = iota
	NodeTypeNormal
	NodeTypeParam
	NodeTypeAll
)

type Node struct{
	nodeType  NodeType
	wildcardParams []string
	childNodesFirstChar [256]*Node
	childNodes []*Node
	matchLen int
	matchChars string
	handler ContextHandler
}

var(
	NOTROOTNODE_ERR = errors.New("only root node can add and find!")
	NODEPATH_ERR = errors.New("node path is error!")
)

func min(a,b int)int{
	if a >= b{
		return b
	}
	return a
}

func (rootNode *Node)insertNode(newNode *Node,path string){
	pathLen := len(path)
	if 0 == len(rootNode.childNodes){
		newNode.matchLen = pathLen
		newNode.matchChars = path
		rootNode.childNodes = append(rootNode.childNodes,newNode)
		rootNode.childNodesFirstChar[byte(newNode.matchChars[0])] = newNode
		return
	}else{
		for _,node := range rootNode.childNodes{
			maxLen := min(node.matchLen,pathLen)
			i := 0
			for ;i<maxLen;i++{
				if node.matchChars[i] != path[i]{
					break
				}
			}
			if 0 == i{
				continue
			}
			if i > 0{
				if node.matchLen >= pathLen{
					unMatchedChars := node.matchChars[i:]
					extendNode := &Node{nodeType:node.nodeType,wildcardParams:node.wildcardParams,
						childNodes:make([]*Node,0),matchLen:len(unMatchedChars),matchChars:unMatchedChars,handler:node.handler}
					if i == pathLen{
						node.nodeType = newNode.nodeType
						node.matchLen = i
						node.matchChars = path
						node.handler = newNode.handler
						extendNode.childNodes = node.childNodes
						node.childNodes = newNode.childNodes
						node.childNodesFirstChar[byte(extendNode.matchChars[0])] = extendNode
						node.childNodes = append(node.childNodes,extendNode)
						newNode.matchChars = path
						newNode.matchLen = i
					}else{
						node.handler = nil
						node.nodeType = NodeTypeNormal
						node.matchLen = i
						node.wildcardParams = nil
						node.matchChars = node.matchChars[0:i]
						newNode.matchChars = path[i:]
						newNode.matchLen = pathLen-i
						node.childNodesFirstChar[byte(newNode.matchChars[0])] = newNode
						node.childNodesFirstChar[byte(extendNode.matchChars[0])] = extendNode 
						node.childNodes = append(node.childNodes,newNode)
						node.childNodes = append(node.childNodes,extendNode)
					}
					return
				}else{
					path = path[i:]
					node.insertNode(newNode,path)//继续往下找
					return
				}
			}
			break
		}
		rootNode.childNodesFirstChar[byte(path[0])] = newNode 
		rootNode.childNodes = append(rootNode.childNodes,newNode)
		newNode.matchChars = path
		newNode.matchLen = pathLen
	}
}

func (rootNode *Node)Add(path string,handler ContextHandler)(error){
	if NodeTypeRoot != rootNode.nodeType{
		return NOTROOTNODE_ERR
	}
	newNode := &Node{nodeType:NodeTypeNormal,childNodes:nil,matchLen:0,matchChars:"",handler:handler}
	pathLen := len(path)
	if pathLen < 2 || path[0] != '/'{
		return NOTROOTNODE_ERR
	}
	if path[pathLen-1] == '/'{
		path = path[0:pathLen-1]
	}
	pathLen = len(path)
	for i:=0;i<pathLen;i++{
		if '*' == path[i] && pathLen > i{
			param := path[i+1:]
			for j:=0;j<len(param);j++{
				if '*' == param[j] || ':' == param[j]{
					return NODEPATH_ERR
				}
			}
			newNode.nodeType = NodeTypeAll
			newNode.wildcardParams=make([]string,0)
			newNode.wildcardParams=append(newNode.wildcardParams,path[i+1:])
			path = path[0:i-1]
			break
		}
		if ':' == path[i] && pathLen > i{
			c := 0
			newNode.wildcardParams=make([]string,0)
			param := path[i+1:]
			j := 0
			for j<len(param){
				if '/' == param[j]{
					j++
					continue
				}
				if '*' == param[j]{
					return NODEPATH_ERR
				}
				if ':' == param[j]{
					newNode.wildcardParams=append(newNode.wildcardParams,param[:j-1])
					c ++
					j ++
					param = param[j:]
				}else{
					j ++
				}
			}
			newNode.wildcardParams=append(newNode.wildcardParams,param[0:])
			newNode.nodeType = NodeTypeParam
			path = path[0:i-1]
			break
		}
	}
	rootNode.insertNode(newNode,path[1:])
	return nil
}

func (nodeRoot *Node)Find(path string,setValue func(key string, value interface{}))(ContextHandler,error){
	if '/' != path[0]{
		return nil,nil
	}
	path = path[1:]
	node := nodeRoot.childNodesFirstChar[byte(path[0])]
	for{
		if nil == node{
			return nil,nil
		}
		if len(path) > node.matchLen{
			if path[:node.matchLen] == node.matchChars {
				path = path[node.matchLen:]
				if NodeTypeNormal == node.nodeType{
					node = node.childNodesFirstChar[byte(path[0])]
					continue
				}
				if '/' == path[0] && len(path) > 1{
					if NodeTypeParam == node.nodeType{
						paramStr := path[1:]
						if 1 == len(node.wildcardParams){
							setValue(node.wildcardParams[0],utils.String2Bytes(paramStr[0:]))
							return node.handler,nil
						}
						i := 0
						j := 0
						for j<len(paramStr){
							if '/' == paramStr[j]{
								if nil != setValue && i < len(node.wildcardParams){
									setValue(node.wildcardParams[i],utils.String2Bytes(paramStr[0:j]))
									i ++
								}
								paramStr = paramStr[j+1:]
								j = 0
								continue
							}
							j++
						}
						if nil != setValue && i < len(node.wildcardParams){
							setValue(node.wildcardParams[i],utils.String2Bytes(paramStr[0:]))
						}
					}else{
					if nil != setValue{
							setValue(node.wildcardParams[0],utils.String2Bytes(path[1:]))
						}
					}
					return node.handler,nil
				}else{
					return nil,nil
				}
			}else{
				return nil,nil
			}
		}else
		if path == node.matchChars{
			return node.handler,nil
		}
		return nil,nil
	}
}