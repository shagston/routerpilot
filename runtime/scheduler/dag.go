package scheduler

import "fmt"

type DAG struct {
	nodes map[string]*dagNode
}

type dagNode struct {
	id       string
	children []string
	parents  []string
}

func NewDAG() *DAG {
	return &DAG{nodes: make(map[string]*dagNode)}
}

func (d *DAG) AddNode(id string) {
	if _, ok := d.nodes[id]; !ok {
		d.nodes[id] = &dagNode{id: id}
	}
}

func (d *DAG) AddEdge(from, to string) error {
	d.AddNode(from)
	d.AddNode(to)

	if d.wouldCycle(from, to) {
		return fmt.Errorf("adding edge %s->%s would create a cycle", from, to)
	}

	d.nodes[from].children = append(d.nodes[from].children, to)
	d.nodes[to].parents = append(d.nodes[to].parents, from)
	return nil
}

func (d *DAG) wouldCycle(from, to string) bool {
	visited := make(map[string]bool)
	return d.reaches(to, from, visited)
}

func (d *DAG) reaches(node, target string, visited map[string]bool) bool {
	if node == target {
		return true
	}
	visited[node] = true
	for _, child := range d.nodes[node].children {
		if !visited[child] && d.reaches(child, target, visited) {
			return true
		}
	}
	return false
}

func (d *DAG) Children(id string) []string {
	if n, ok := d.nodes[id]; ok {
		return n.children
	}
	return nil
}

func (d *DAG) Parents(id string) []string {
	if n, ok := d.nodes[id]; ok {
		return n.parents
	}
	return nil
}

func (d *DAG) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for id := range d.nodes {
		inDegree[id] = 0
	}
	for _, n := range d.nodes {
		for _, child := range n.children {
			inDegree[child]++
		}
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var result []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, id)

		for _, child := range d.nodes[id].children {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(result) != len(d.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}
