package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Graph struct {
	// The order of node and edge doesn't matter.
	Nodes map[string]*Node // key is Node.ID
	Edges []*Edge
}

type Node struct {
	ID        string
	Label     string
	ParentIDs map[string]struct{}
	ChildIDs  map[string]struct{}
}

type Edge struct {
	From string
	To   string
}

func main() {
}

func parseModGraph(r io.Reader) (*Graph, error) {
	nodes := map[string]*Node{}
	edges := []*Edge{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		parentName, childName, err := getParentAndChildName(line)
		if err != nil {
			return nil, fmt.Errorf("the content contains a invalid line: %w", err)
		}

		parentNode, ok := nodes[parentName]
		if !ok {
			// Initialize a parent node.
			parentNode = &Node{
				ID:        parentName,
				Label:     parentName,
				ParentIDs: map[string]struct{}{},
				ChildIDs:  map[string]struct{}{},
			}
			nodes[parentName] = parentNode
		}
		parentNode.ChildIDs[childName] = struct{}{}

		childNode, ok := nodes[childName]
		if !ok {
			// Initialize a child node.
			childNode = &Node{
				ID:        childName,
				Label:     childName,
				ParentIDs: map[string]struct{}{},
				ChildIDs:  map[string]struct{}{},
			}
			nodes[childName] = childNode
		}
		childNode.ParentIDs[parentName] = struct{}{}

		edges = append(edges, &Edge{From: parentName, To: childName})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan the content: %w", err)
	}

	return &Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

var validLineRegexp = regexp.MustCompile(`^[a-zA-Z]`)

func getParentAndChildName(line string) (parent, child string, err error) {
	ss := strings.Split(line, " ")
	if len(ss) < 2 {
		return "", "", fmt.Errorf("a line must have two elements separated by white space: '%s'", line)
	}
	parent = ss[0]
	child = ss[1]
	if !validLineRegexp.MatchString(parent) || !validLineRegexp.MatchString(child) {
		return "", "", fmt.Errorf("elements of parent or child must start with a letter of the alphabet: '%s'", line)
	}
	return
}
