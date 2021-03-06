package gee

import "strings"

//前缀树是用来匹配pattern
//来了一个请求，用前缀树可以匹配到一个pattern上
//再通过这个pattern，用map找到对应的handler
type node struct {
	pattern  string  //待匹配路由，例如/p/:lang
	part     string  //路由中的一部分，例如:lang
	children []*node //子节点，例如[doc,tutorial,intro]
	isWild   bool    //是否精确匹配
}

//第一个匹配成功的节点用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

//所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		//只在最后一个节点赋值pattern
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part) //找第一个符合part的节点
	if child == nil {           //找不到就创建一个
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
