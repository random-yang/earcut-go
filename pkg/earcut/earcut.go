package earcut

import "math"

// Triangulate 将二维多边形进行三角剖分
// vertices: 多边形顶点坐标数组，格式为 [x1, y1, x2, y2, ...]
// holes: 多边形中的洞，每个元素表示洞的起始顶点在 vertices 中的索引
// dim: 坐标维度，对于二维坐标为 2
func Triangulate(vertices []float64, holes []int, dim int) []int {
	hasHoles := holes != nil && len(holes) > 0
	outerLen := len(vertices)

	if len(vertices) < 6 {
		return nil
	}

	// 创建外圈链表
	outerNode := linkedList(vertices, 0, outerLen, dim, true)
	if outerNode == nil || outerNode.next == outerNode.prev {
		return nil
	}

	var minX, minY, maxX, maxY float64
	var invSize float64

	// 如果有洞，添加洞到外圈链表中
	if hasHoles {
		outerNode = eliminateHoles(vertices, holes, outerNode, dim)
	}

	// 如果外圈是逆时针的，需要翻转为顺时针
	if signedArea(vertices, 0, outerLen, dim) > 0 {
		outerNode = outerNode.next
	}

	// 计算边界框并初始化 z-order 曲线
	minX = vertices[0]
	minY = vertices[1]
	maxX = minX
	maxY = minY
	node := outerNode
	for {
		if node.x < minX {
			minX = node.x
		}
		if node.y < minY {
			minY = node.y
		}
		if node.x > maxX {
			maxX = node.x
		}
		if node.y > maxY {
			maxY = node.y
		}
		node = node.next
		if node == outerNode {
			break
		}
	}

	// 计算 z-order 曲线值的缩放因子
	invSize = math.Max(maxX-minX, maxY-minY)
	if invSize != 0 {
		invSize = 32767 / invSize
	}

	// 三角剖分过程
	return earcutLinked(outerNode, []int{}, dim, minX, minY, invSize)
}

// linkedList 创建一个顶点的双向链表
func linkedList(data []float64, start, end, dim int, clockwise bool) *Node {
	if len(data) < 4 {
		return nil
	}

	var last *Node
	if clockwise == (signedArea(data, start, end, dim) > 0) {
		for i := end - dim; i >= start; i -= dim {
			last = insertNode(i/dim, data[i], data[i+1], last)
		}
	} else {
		for i := start; i < end; i += dim {
			last = insertNode(i/dim, data[i], data[i+1], last)
		}
	}

	if last != nil && equals(last, last.next) {
		removeNode(last)
		last = last.next
	}

	return last
}

// eliminateHoles 将洞连接到外圈
func eliminateHoles(data []float64, holes []int, outerNode *Node, dim int) *Node {
	var queue []*Node

	// 将每个洞的最左边的顶点加入队列
	for i := 0; i < len(holes); i++ {
		start := holes[i] * dim
		var end int
		if i < len(holes)-1 {
			end = holes[i+1] * dim
		} else {
			end = len(data)
		}
		list := linkedList(data, start, end, dim, false)
		if list == list.next {
			list.steiner = true
		}
		queue = append(queue, getLeftmost(list))
	}

	// 按照 y 坐标排序
	sort(queue)

	// 逐个处理洞
	for i := 0; i < len(queue); i++ {
		eliminateHole(queue[i], outerNode)
		outerNode = filterPoints(outerNode, outerNode.next)
	}

	return outerNode
}

// earcutLinked 主要的耳切算法实现
func earcutLinked(ear *Node, triangles []int, dim int, minX, minY, invSize float64) []int {
	if ear == nil {
		return triangles
	}

	// 如果整个多边形退化为一个三角形
	if ear.prev == ear.next.next {
		triangles = append(triangles, ear.prev.i)
		triangles = append(triangles, ear.i)
		triangles = append(triangles, ear.next.i)
		return triangles
	}

	stop := ear
	var prev, next *Node

	// 遍历耳朵，直到找到一个有效的或者回到起点
	for ear != nil && ear.prev != ear.next {
		prev = ear.prev
		next = ear.next

		if isValidEar(ear) {
			// 切掉耳朵
			triangles = append(triangles, prev.i)
			triangles = append(triangles, ear.i)
			triangles = append(triangles, next.i)

			// 从链表中移除ear节点
			removeNode(ear)

			// 跳过下一个顶点，因为它现在是一个耳朵
			ear = next.next
			stop = next.next

			continue
		}

		ear = next
		if ear == stop {
			// 如果到达停止点，说明无法找到更多的耳朵
			// 在这种情况下，我们通过过滤点来尝试再次处理
			if len(triangles) == 0 {
				ear = filterPoints(ear, nil)
				if ear == nil {
					break
				}
				stop = ear
			} else {
				break
			}
		}
	}

	return triangles
}

// isValidEar 检查一个三角形是否是一个有效的耳朵
func isValidEar(ear *Node) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if area(a, b, c) >= 0 {
		return false
	}

	// 检查是否有其他点在三角形内部
	node := ear.next.next
	for node != ear.prev {
		if pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, node.x, node.y) &&
			area(node.prev, node, node.next) >= 0 {
			return false
		}
		node = node.next
	}

	return true
}

// area 计算三角形面积
func area(p, q, r *Node) float64 {
	return (q.y-p.y)*(r.x-q.x) - (q.x-p.x)*(r.y-q.y)
}

// pointInTriangle 判断点是否在三角形内部
func pointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	return (cx-px)*(ay-py)-(ax-px)*(cy-py) >= 0 &&
		(ax-px)*(by-py)-(bx-px)*(ay-py) >= 0 &&
		(bx-px)*(cy-py)-(cx-px)*(by-py) >= 0
}

// getLeftmost 获取最左边的顶点
func getLeftmost(start *Node) *Node {
	node := start
	leftmost := start
	for {
		if node.x < leftmost.x || (node.x == leftmost.x && node.y < leftmost.y) {
			leftmost = node
		}
		node = node.next
		if node == start {
			break
		}
	}
	return leftmost
}

// filterPoints 移除重复的点
func filterPoints(start, end *Node) *Node {
	if start == nil {
		return nil
	}
	if end == nil {
		end = start
	}

	p := start
	for {
		again := false

		if !p.steiner && equals(p, p.next) || area(p.prev, p, p.next) == 0 {
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

// equals 判断两个点是否相等
func equals(p1, p2 *Node) bool {
	return p1.x == p2.x && p1.y == p2.y
}

// sort 对节点数组按照 y 坐标排序
func sort(nodes []*Node) {
	for i := 1; i < len(nodes); i++ {
		j := i
		temp := nodes[i]
		for j > 0 && less(temp, nodes[j-1]) {
			nodes[j] = nodes[j-1]
			j--
		}
		nodes[j] = temp
	}
}

// less 比较两个节点的位置
func less(a, b *Node) bool {
	return a.y < b.y
}

// Point 表示二维平面上的一个点
type Point struct {
	X, Y float64
}

// Triangle 表示一个三角形
type Triangle struct {
	A, B, C Point
}
