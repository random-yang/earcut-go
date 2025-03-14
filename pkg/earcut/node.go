package earcut

// Node 表示多边形中的一个节点
type Node struct {
	// 节点坐标
	x, y float64

	// 节点在原始数据中的索引
	i int

	// 顶点的类型（是否为外圈或内圈的起始点）
	steiner bool

	// 双向链表指针
	prev *Node
	next *Node

	// z坐标，用于存储顶点的签名值
	z int

	// 前一个和后一个Z值
	prevZ *Node
	nextZ *Node

	// 输入数据是否已被处理
	visited bool
}

// createNode 创建一个新的节点
func createNode(i int, x, y float64) *Node {
	return &Node{
		i:       i,
		x:       x,
		y:       y,
		steiner: false,
	}
}

// signedArea 计算多边形的有符号面积
func signedArea(data []float64, start, end int, dim int) float64 {
	var sum float64
	j := end - dim
	for i := start; i < end; i += dim {
		sum += (data[j] - data[i]) * (data[i+1] + data[j+1])
		j = i
	}
	return sum
}

// insertNode 在指定节点后插入新节点
func insertNode(i int, x, y float64, last *Node) *Node {
	p := createNode(i, x, y)

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

// removeNode 从链表中移除节点
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
