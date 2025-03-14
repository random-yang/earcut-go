// Package earcut implements a fast and tiny JavaScript polygon triangulation library.
// This is a Go port of the JavaScript library: https://github.com/mapbox/earcut
package earcut

import (
	"math"
	"sort"
)

// Earcut triangulates the given polygon with optional holes.
// data is a flat array of vertex coordinates like [x0,y0, x1,y1, x2,y2, ...].
// holeIndices is an array of hole indices if any (e.g. [5, 8] for a 12-vertex input would mean
// one hole with vertices 5–7 and another with 8–11).
// dim is the number of coordinates per vertex in the input array (2 by default).
// Returns a flat array of triangle indices like [a,b,c, d,e,f, ...].
func Earcut(data []float64, holeIndices []int, dim int) []int {
	if dim == 0 {
		dim = 2
	}

	hasHoles := holeIndices != nil && len(holeIndices) > 0
	outerLen := 0
	if hasHoles {
		outerLen = holeIndices[0] * dim
	} else {
		outerLen = len(data)
	}

	outerNode := linkedList(data, 0, outerLen, dim, true)
	triangles := []int{}

	if outerNode == nil || outerNode.next == outerNode.prev {
		return triangles
	}

	var minX, minY, invSize float64

	if hasHoles {
		outerNode = eliminateHoles(data, holeIndices, outerNode, dim)
	}

	// if the shape is not too simple, we'll use z-order curve hash later; calculate polygon bbox
	if len(data) > 80*dim {
		minX = data[0]
		minY = data[1]
		maxX := data[0]
		maxY := data[1]

		for i := dim; i < outerLen; i += dim {
			x := data[i]
			y := data[i+1]
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}

		// minX, minY and invSize are later used to transform coords into integers for z-order calculation
		invSize = 0
		if maxX-minX > maxY-minY {
			invSize = 32767 / (maxX - minX)
		} else {
			invSize = 32767 / (maxY - minY)
		}
	}

	earcutLinked(outerNode, &triangles, dim, minX, minY, invSize, 0)

	return triangles
}

// Node represents a vertex in a doubly-linked list
type Node struct {
	// vertex index in coordinates array
	i int
	// vertex coordinates
	x, y float64
	// previous and next vertex nodes in a polygon ring
	prev, next *Node
	// z-order curve value
	z int
	// previous and next nodes in z-order
	prevZ, nextZ *Node
	// indicates whether this is a steiner point
	steiner bool
}

// create a circular doubly linked list from polygon points in the specified winding order
func linkedList(data []float64, start, end, dim int, clockwise bool) *Node {
	var last *Node

	if clockwise == (signedArea(data, start, end, dim) > 0) {
		for i := start; i < end; i += dim {
			last = insertNode(i/dim, data[i], data[i+1], last)
		}
	} else {
		for i := end - dim; i >= start; i -= dim {
			last = insertNode(i/dim, data[i], data[i+1], last)
		}
	}

	if last != nil && equals(last, last.next) {
		removeNode(last)
		last = last.next
	}

	return last
}

// create a node and optionally link it with previous one (in a circular doubly linked list)
func insertNode(i int, x, y float64, last *Node) *Node {
	p := &Node{
		i: i,
		x: x,
		y: y,
	}

	if last == nil {
		p.prev = p
		p.next = p
	} else {
		p.next = last.next
		p.prev = last
		last.next.prev = p
		last.next = p
	}
	return p
}

func removeNode(p *Node) {
	p.next.prev = p.prev
	p.prev.next = p.next

	if p.prevZ != nil {
		p.prevZ.nextZ = p.nextZ
	}
	if p.nextZ != nil {
		p.nextZ.prevZ = p.prevZ
	}
}

// signed area of a triangle
func area(p, q, r *Node) float64 {
	return (q.y-p.y)*(r.x-q.x) - (q.x-p.x)*(r.y-q.y)
}

// check if two points are equal
func equals(p1, p2 *Node) bool {
	return p1.x == p2.x && p1.y == p2.y
}

// signed area of a polygon
func signedArea(data []float64, start, end, dim int) float64 {
	var sum float64
	for i, j := start, end-dim; i < end; i += dim {
		sum += (data[j] - data[i]) * (data[i+1] + data[j+1])
		j = i
	}
	return sum
}

// eliminate colinear or duplicate points
func filterPoints(start, end *Node) *Node {
	if start == nil {
		return start
	}
	if end == nil {
		end = start
	}

	p := start
	var again bool

	for {
		again = false

		if !p.steiner && (equals(p, p.next) || area(p.prev, p, p.next) == 0) {
			removeNode(p)
			p = end
			p = p.prev
			if p == p.next {
				break
			}
			again = true
		} else {
			p = p.next
		}

		if !again && p == end {
			break
		}
	}

	return end
}

// main ear slicing loop which triangulates a polygon (given as a linked list)
func earcutLinked(ear *Node, triangles *[]int, dim int, minX, minY, invSize float64, pass int) {
	if ear == nil {
		return
	}

	// interlink polygon nodes in z-order
	if pass == 0 && invSize != 0 {
		indexCurve(ear, minX, minY, invSize)
	}

	stop := ear

	// iterate through ears, slicing them one by one
	for ear.prev != ear.next {
		prev := ear.prev
		next := ear.next

		var isEarValid bool
		if invSize != 0 {
			isEarValid = isEarHashed(ear, minX, minY, invSize)
		} else {
			isEarValid = isEar(ear)
		}

		if isEarValid {
			// cut off the triangle
			*triangles = append(*triangles, prev.i, ear.i, next.i)

			removeNode(ear)

			// skipping the next vertex leads to less sliver triangles
			ear = next.next
			stop = next.next

			continue
		}

		ear = next

		// if we looped through the whole remaining polygon and can't find any more ears
		if ear == stop {
			// try filtering points and slicing again
			if pass == 0 {
				earcutLinked(filterPoints(ear, nil), triangles, dim, minX, minY, invSize, 1)
			} else if pass == 1 {
				// if this didn't work, try curing all small self-intersections locally
				ear = cureLocalIntersections(filterPoints(ear, nil), triangles)
				earcutLinked(ear, triangles, dim, minX, minY, invSize, 2)
			} else if pass == 2 {
				// as a last resort, try splitting the remaining polygon into two
				splitEarcut(ear, triangles, dim, minX, minY, invSize)
			}
			break
		}
	}
}

// check whether a polygon node forms a valid ear with adjacent nodes
func isEar(ear *Node) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if area(a, b, c) >= 0 {
		return false // reflex, can't be an ear
	}

	// now make sure we don't have other points inside the potential ear
	ax, bx, cx := a.x, b.x, c.x
	ay, by, cy := a.y, b.y, c.y

	// triangle bbox
	x0 := min3(ax, bx, cx)
	y0 := min3(ay, by, cy)
	x1 := max3(ax, bx, cx)
	y1 := max3(ay, by, cy)

	p := c.next
	for p != a {
		if p.x >= x0 && p.x <= x1 && p.y >= y0 && p.y <= y1 &&
			pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, p.x, p.y) &&
			area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.next
	}

	return true
}

func min3(a, b, c float64) float64 {
	return min(min(a, b), c)
}

func max3(a, b, c float64) float64 {
	return max(max(a, b), c)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// check if a point lies within a convex triangle
func pointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	return (cx-px)*(ay-py) >= (ax-px)*(cy-py) &&
		(ax-px)*(by-py) >= (bx-px)*(ay-py) &&
		(bx-px)*(cy-py) >= (cx-px)*(by-py)
}

// check if a point lies within a convex triangle but false if its equal to the first point of the triangle
func pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, px, py float64) bool {
	return !(ax == px && ay == py) && pointInTriangle(ax, ay, bx, by, cx, cy, px, py)
}

// check whether a polygon node forms a valid ear with adjacent nodes
func isEarHashed(ear *Node, minX, minY, invSize float64) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if area(a, b, c) >= 0 {
		return false // reflex, can't be an ear
	}

	ax, bx, cx := a.x, b.x, c.x
	ay, by, cy := a.y, b.y, c.y

	// triangle bbox
	x0 := min3(ax, bx, cx)
	y0 := min3(ay, by, cy)
	x1 := max3(ax, bx, cx)
	y1 := max3(ay, by, cy)

	// z-order range for the current triangle bbox
	minZ := zOrder(x0, y0, minX, minY, invSize)
	maxZ := zOrder(x1, y1, minX, minY, invSize)

	p := ear.prevZ
	n := ear.nextZ

	// look for points inside the triangle in both directions
	for p != nil && p.z >= minZ && n != nil && n.z <= maxZ {
		if p.x >= x0 && p.x <= x1 && p.y >= y0 && p.y <= y1 && p != a && p != c &&
			pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, p.x, p.y) && area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.prevZ

		if n.x >= x0 && n.x <= x1 && n.y >= y0 && n.y <= y1 && n != a && n != c &&
			pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, n.x, n.y) && area(n.prev, n, n.next) >= 0 {
			return false
		}
		n = n.nextZ
	}

	// look for remaining points in decreasing z-order
	for p != nil && p.z >= minZ {
		if p.x >= x0 && p.x <= x1 && p.y >= y0 && p.y <= y1 && p != a && p != c &&
			pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, p.x, p.y) && area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.prevZ
	}

	// look for remaining points in increasing z-order
	for n != nil && n.z <= maxZ {
		if n.x >= x0 && n.x <= x1 && n.y >= y0 && n.y <= y1 && n != a && n != c &&
			pointInTriangleExceptFirst(ax, ay, bx, by, cx, cy, n.x, n.y) && area(n.prev, n, n.next) >= 0 {
			return false
		}
		n = n.nextZ
	}

	return true
}

// go through all polygon nodes and cure small local self-intersections
func cureLocalIntersections(start *Node, triangles *[]int) *Node {
	p := start
	for {
		a := p.prev
		b := p.next.next

		if !equals(a, b) && intersects(a, p, p.next, b) && locallyInside(a, b) && locallyInside(b, a) {
			*triangles = append(*triangles, a.i, p.i, b.i)

			// remove two nodes involved
			removeNode(p)
			removeNode(p.next)

			p = b
			start = b
		}
		p = p.next
		if p == start {
			break
		}
	}

	return filterPoints(p, nil)
}

// try splitting polygon into two and triangulate them independently
func splitEarcut(start *Node, triangles *[]int, dim int, minX, minY, invSize float64) {
	// look for a valid diagonal that divides the polygon into two
	a := start
	for {
		b := a.next.next
		for b != a.prev {
			if a.i != b.i && isValidDiagonal(a, b) {
				// split the polygon in two by the diagonal
				c := splitPolygon(a, b)

				// filter colinear points around the cuts
				a = filterPoints(a, a.next)
				c = filterPoints(c, c.next)

				// run earcut on each half
				earcutLinked(a, triangles, dim, minX, minY, invSize, 0)
				earcutLinked(c, triangles, dim, minX, minY, invSize, 0)
				return
			}
			b = b.next
		}
		a = a.next
		if a == start {
			break
		}
	}
}

// link every hole into the outer loop, producing a single-ring polygon without holes
func eliminateHoles(data []float64, holeIndices []int, outerNode *Node, dim int) *Node {
	queue := []*Node{}

	for i, length := 0, len(holeIndices); i < length; i++ {
		start := holeIndices[i] * dim
		end := 0
		if i < length-1 {
			end = holeIndices[i+1] * dim
		} else {
			end = len(data)
		}
		list := linkedList(data, start, end, dim, false)
		if list == list.next {
			list.steiner = true
		}
		queue = append(queue, getLeftmost(list))
	}

	// sort holes from left to right
	sortByXYSlope(queue)

	// process holes from left to right
	for i := 0; i < len(queue); i++ {
		outerNode = eliminateHole(queue[i], outerNode)
	}

	return outerNode
}

// sort an array of nodes by x, then y, then slope
func sortByXYSlope(nodes []*Node) {
	// Sort using Go's sort package
	sort.Slice(nodes, func(i, j int) bool {
		a, b := nodes[i], nodes[j]
		if a.x != b.x {
			return a.x < b.x
		}
		if a.y != b.y {
			return a.y < b.y
		}
		// when two holes' leftmost points are at the same vertex, sort counterclockwise
		aSlope := (a.next.y - a.y) / (a.next.x - a.x)
		bSlope := (b.next.y - b.y) / (b.next.x - b.x)
		return aSlope < bSlope
	})
}

// find a bridge between vertices that connects hole with an outer ring and link it
func eliminateHole(hole, outerNode *Node) *Node {
	bridge := findHoleBridge(hole, outerNode)
	if bridge == nil {
		return outerNode
	}

	bridgeReverse := splitPolygon(bridge, hole)

	// filter collinear points around the cuts
	filterPoints(bridgeReverse, bridgeReverse.next)
	return filterPoints(bridge, bridge.next)
}

// David Eberly's algorithm for finding a bridge between hole and outer polygon
func findHoleBridge(hole, outerNode *Node) *Node {
	p := outerNode
	hx := hole.x
	hy := hole.y
	qx := math.Inf(-1)
	var m *Node

	// find a segment intersected by a ray from the hole's leftmost point to the left;
	// segment's endpoint with lesser x will be potential connection point
	// unless they intersect at a vertex, then choose the vertex
	if equals(hole, p) {
		return p
	}
	for {
		if equals(hole, p.next) {
			return p.next
		} else if hy <= p.y && hy >= p.next.y && p.next.y != p.y {
			x := p.x + (hy-p.y)*(p.next.x-p.x)/(p.next.y-p.y)
			if x <= hx && x > qx {
				qx = x
				if p.x < p.next.x {
					m = p
				} else {
					m = p.next
				}
				if x == hx {
					return m // hole touches outer segment; pick leftmost endpoint
				}
			}
		}
		p = p.next
		if p == outerNode {
			break
		}
	}

	if m == nil {
		return nil
	}

	// look for points inside the triangle of hole point, segment intersection and endpoint;
	// if there are no points found, we have a valid connection;
	// otherwise choose the point of the minimum angle with the ray as connection point

	stop := m
	mx := m.x
	my := m.y
	tanMin := math.Inf(1)

	p = m

	for {
		if hx >= p.x && p.x >= mx && hx != p.x &&
			pointInTriangle(hx, hy, qx, hy, mx, my, p.x, p.y) {

			tan := math.Abs(hy-p.y) / (hx - p.x) // tangential

			if locallyInside(p, hole) &&
				(tan < tanMin || (tan == tanMin && (p.x > m.x || (p.x == m.x && sectorContainsSector(m, p))))) {
				m = p
				tanMin = tan
			}
		}

		p = p.next
		if p == stop {
			break
		}
	}

	return m
}

// whether sector in vertex m contains sector in vertex p in the same coordinates
func sectorContainsSector(m, p *Node) bool {
	return area(m.prev, m, p.prev) < 0 && area(p.next, m, m.next) < 0
}

// interlink polygon nodes in z-order
func indexCurve(start *Node, minX, minY, invSize float64) {
	p := start
	for {
		if p.z == 0 {
			p.z = int(zOrder(p.x, p.y, minX, minY, invSize))
		}
		p.prevZ = p.prev
		p.nextZ = p.next
		p = p.next
		if p == start {
			break
		}
	}

	p.prevZ.nextZ = nil
	p.prevZ = nil

	sortLinked(p)
}

// Simon Tatham's linked list merge sort algorithm
// http://www.chiark.greenend.org.uk/~sgtatham/algorithms/listsort.html
func sortLinked(list *Node) *Node {
	var inSize int = 1
	var numMerges int

	for {
		p := list
		list = nil
		var tail *Node = nil
		numMerges = 0

		for p != nil {
			numMerges++
			q := p
			pSize := 0
			for i := 0; i < inSize; i++ {
				pSize++
				q = q.nextZ
				if q == nil {
					break
				}
			}
			qSize := inSize

			for pSize > 0 || (qSize > 0 && q != nil) {
				var e *Node

				if pSize != 0 && (qSize == 0 || q == nil || p.z <= q.z) {
					e = p
					p = p.nextZ
					pSize--
				} else {
					e = q
					q = q.nextZ
					qSize--
				}

				if tail != nil {
					tail.nextZ = e
				} else {
					list = e
				}

				e.prevZ = tail
				tail = e
			}

			p = q
		}

		tail.nextZ = nil
		inSize *= 2

		if numMerges <= 1 {
			break
		}
	}

	return list
}

// z-order of a point given coords and inverse of the longer side of data bbox
func zOrder(x, y, minX, minY, invSize float64) int {
	// coords are transformed into non-negative 15-bit integer range
	ix := int((x - minX) * invSize)
	iy := int((y - minY) * invSize)

	ix = (ix | (ix << 8)) & 0x00FF00FF
	ix = (ix | (ix << 4)) & 0x0F0F0F0F
	ix = (ix | (ix << 2)) & 0x33333333
	ix = (ix | (ix << 1)) & 0x55555555

	iy = (iy | (iy << 8)) & 0x00FF00FF
	iy = (iy | (iy << 4)) & 0x0F0F0F0F
	iy = (iy | (iy << 2)) & 0x33333333
	iy = (iy | (iy << 1)) & 0x55555555

	return ix | (iy << 1)
}

// find the leftmost node of a polygon ring
func getLeftmost(start *Node) *Node {
	p := start
	leftmost := start
	for {
		if p.x < leftmost.x || (p.x == leftmost.x && p.y < leftmost.y) {
			leftmost = p
		}
		p = p.next
		if p == start {
			break
		}
	}
	return leftmost
}

// check if a diagonal between two polygon nodes is valid (lies in polygon interior)
func isValidDiagonal(a, b *Node) bool {
	return a.next.i != b.i && a.prev.i != b.i && !intersectsPolygon(a, b) && // doesn't intersect other edges
		(locallyInside(a, b) && locallyInside(b, a) && middleInside(a, b) && // locally visible
			(area(a.prev, a, b.prev) != 0 || area(a, b.prev, b) != 0) || // does not create opposite-facing sectors
			equals(a, b) && area(a.prev, a, a.next) > 0 && area(b.prev, b, b.next) > 0) // special zero-length case
}

// check if two segments intersect
func intersects(p1, q1, p2, q2 *Node) bool {
	o1 := sign(area(p1, q1, p2))
	o2 := sign(area(p1, q1, q2))
	o3 := sign(area(p2, q2, p1))
	o4 := sign(area(p2, q2, q1))

	if o1 != o2 && o3 != o4 {
		return true // general case
	}

	if o1 == 0 && onSegment(p1, p2, q1) {
		return true // p1, q1 and p2 are collinear and p2 lies on p1q1
	}
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true // p1, q1 and q2 are collinear and q2 lies on p1q1
	}
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true // p2, q2 and p1 are collinear and p1 lies on p2q2
	}
	if o4 == 0 && onSegment(p2, q1, q2) {
		return true // p2, q2 and q1 are collinear and q1 lies on p2q2
	}

	return false
}

// for collinear points p, q, r, check if point q lies on segment pr
func onSegment(p, q, r *Node) bool {
	return q.x <= math.Max(p.x, r.x) && q.x >= math.Min(p.x, r.x) &&
		q.y <= math.Max(p.y, r.y) && q.y >= math.Min(p.y, r.y)
}

func sign(num float64) int {
	if num > 0 {
		return 1
	}
	if num < 0 {
		return -1
	}
	return 0
}

// check if a polygon diagonal intersects any polygon segments
func intersectsPolygon(a, b *Node) bool {
	p := a
	for {
		if p.i != a.i && p.next.i != a.i && p.i != b.i && p.next.i != b.i &&
			intersects(p, p.next, a, b) {
			return true
		}
		p = p.next
		if p == a {
			break
		}
	}
	return false
}

// check if a polygon diagonal is locally inside the polygon
func locallyInside(a, b *Node) bool {
	if area(a.prev, a, a.next) < 0 {
		return area(a, b, a.next) >= 0 && area(a, a.prev, b) >= 0
	}
	return area(a, b, a.prev) < 0 || area(a, a.next, b) < 0
}

// check if the middle point of a polygon diagonal is inside the polygon
func middleInside(a, b *Node) bool {
	p := a
	inside := false
	px := (a.x + b.x) / 2
	py := (a.y + b.y) / 2
	for {
		if ((p.y > py) != (p.next.y > py)) && p.next.y != p.y &&
			(px < (p.next.x-p.x)*(py-p.y)/(p.next.y-p.y)+p.x) {
			inside = !inside
		}
		p = p.next
		if p == a {
			break
		}
	}
	return inside
}

// link two polygon vertices with a bridge; if the vertices belong to the same ring, it splits polygon into two;
// if one belongs to the outer ring and another to a hole, it merges it into a single ring
func splitPolygon(a, b *Node) *Node {
	a2 := createNode(a.i, a.x, a.y)
	b2 := createNode(b.i, b.x, b.y)
	an := a.next
	bp := b.prev

	a.next = b
	b.prev = a

	a2.next = an
	an.prev = a2

	b2.next = a2
	a2.prev = b2

	bp.next = b2
	b2.prev = bp

	return b2
}

func createNode(i int, x, y float64) *Node {
	return &Node{
		i:       i,
		x:       x,
		y:       y,
		prev:    nil,
		next:    nil,
		z:       0,
		prevZ:   nil,
		nextZ:   nil,
		steiner: false,
	}
}

// Deviation returns a percentage difference between the polygon area and its triangulation area;
// used to verify correctness of triangulation
func Deviation(data []float64, holeIndices []int, dim int, triangles []int) float64 {
	hasHoles := holeIndices != nil && len(holeIndices) > 0
	outerLen := 0
	if hasHoles {
		outerLen = holeIndices[0] * dim
	} else {
		outerLen = len(data)
	}

	polygonArea := math.Abs(signedArea(data, 0, outerLen, dim))
	if hasHoles {
		for i, length := 0, len(holeIndices); i < length; i++ {
			start := holeIndices[i] * dim
			end := 0
			if i < length-1 {
				end = holeIndices[i+1] * dim
			} else {
				end = len(data)
			}
			polygonArea -= math.Abs(signedArea(data, start, end, dim))
		}
	}

	var trianglesArea float64
	for i := 0; i < len(triangles); i += 3 {
		a := triangles[i] * dim
		b := triangles[i+1] * dim
		c := triangles[i+2] * dim
		trianglesArea += math.Abs(
			(data[a]-data[c])*(data[b+1]-data[a+1]) -
				(data[a]-data[b])*(data[c+1]-data[a+1]))
	}

	if polygonArea == 0 && trianglesArea == 0 {
		return 0
	}
	return math.Abs((trianglesArea - polygonArea) / polygonArea)
}

// Flatten turns a polygon in a multi-dimensional array form (e.g. as in GeoJSON) into a form Earcut accepts
func Flatten(data [][][]float64) (vertices []float64, holes []int, dim int) {
	if len(data) == 0 {
		return nil, nil, 0
	}

	dim = len(data[0][0])
	holeIndex := 0
	prevLen := 0

	for i, ring := range data {
		for _, p := range ring {
			for _, coord := range p {
				vertices = append(vertices, coord)
			}
		}
		if i > 0 {
			holeIndex += prevLen
			holes = append(holes, holeIndex)
		}
		prevLen = len(ring)
	}

	return vertices, holes, dim
}

// Triangulate is an alias for Earcut for backward compatibility
func Triangulate(data []float64, holeIndices []int, dim int) []int {
	return Earcut(data, holeIndices, dim)
}
