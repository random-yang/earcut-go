package earcut

import "math"

// eliminateHole 将一个洞连接到外部多边形
func eliminateHole(hole, outerNode *Node) {
	bridge := findHoleBridge(hole, outerNode)
	if bridge == nil {
		return
	}

	bridgeReverse := splitPolygon(bridge, hole)
	filterPoints(bridgeReverse, bridgeReverse.next)
}

// findHoleBridge 找到连接洞和外部多边形的桥
func findHoleBridge(hole, outerNode *Node) *Node {
	p := outerNode
	hx := hole.x
	hy := hole.y
	qx := math.Inf(-1)
	var m *Node

	// 找到位于洞右侧的点，选择最接近的一个
	for {
		if hy <= p.y && hy >= p.next.y && p.next.y != p.y {
			x := p.x + (hy-p.y)*(p.next.x-p.x)/(p.next.y-p.y)
			if x <= hx && x > qx {
				qx = x
				if x == hx {
					if hy == p.y {
						return p
					}
					if hy == p.next.y {
						return p.next
					}
				}
				m = p
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

	// 检查是否在正确的位置
	if hx == qx {
		return m.prev
	}

	// 寻找最近的点
	p = m
	for {
		if p.x >= hx && pointInTriangle(if_lt(qx, hx, p.x, qx), hy,
			if_lt(qx, hx, p.next.x, qx), p.next.y,
			if_lt(qx, hx, p.prev.x, qx), p.prev.y, hx, hy) {
			return p
		}
		p = p.next
		if p == m {
			break
		}
	}

	return nil
}

// splitPolygon 在两个点之间分割多边形
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

// if_lt 如果 x < y 返回 a，否则返回 b
func if_lt(x, y, a, b float64) float64 {
	if x < y {
		return a
	}
	return b
}
