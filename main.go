package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const helpMsg = `A tool for interactively viewing a dependency graph of go module.

Usage: go mod graph | modgrapher OR modgrapher [file]

[file] in args must be a result of 'go mod graph' command output.
`

var fs = flag.NewFlagSet("modgrapher", flag.ExitOnError)

func init() {
	fs.Usage = func() { fmt.Fprintln(os.Stderr, helpMsg) }
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(1)
	}
}

func run(args []string) (err error) {
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse command args: %w", err)
	}

	if fs.NArg() > 1 {
		return errors.New("too many arguments")
	}

	filename := fs.Arg(0)
	var reader io.Reader
	switch filename {
	case "", "-":
		reader = os.Stdin
	default:
		// If a file to be read is specified as an argument
		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		defer func() {
			if err = f.Close(); err != nil {
				err = fmt.Errorf("%w", err)
			}
		}()
		reader = f
	}

	graph, err := parseModGraph(reader)
	if err != nil {
		return fmt.Errorf("failed to parse input: %w", err)
	}

	fmt.Printf("%+v\n", *graph)

	return nil
}

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

func parseModGraph(r io.Reader) (*Graph, error) {
	nodes := map[string]*Node{}
	edges := []*Edge{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		parentName, childName, err := getParentAndChildName(line)
		if err != nil {
			return nil, fmt.Errorf("input content contains a invalid line: %w", err)
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
