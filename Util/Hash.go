package Util

import (
    "crypto/sha1"
    "fmt"
    "hash/crc32"
    "io"
    "math"
    "sort"
)

func HashSha1Crc32(s string) uint32 {
    sh := sha1.New()
    io.WriteString(sh, s)
    return crc32.ChecksumIEEE(sh.Sum(nil))
}

// new hash ring, optional args:
// node string: node item key, multi node accepted, see AddNode().
// spots int: num of virtual spots for each node, default 32.
// hashFunc HashFunc: function to calculate hash value, default HashSha1Crc32.
// eg. h := NewHashRing("127.0.0.1:6379", "127.0.0.1:6379", "127.0.0.1:6380")
func NewHashRing(args ...interface{}) *HashRing {
    h := &HashRing{
        numSpots: 32,
        hashFunc: HashSha1Crc32,
        weights:  make(map[string]int),
        items:    make([]hashItem, 0),
    }

    for _, arg := range args {
        switch v := arg.(type) {
        case string:
            if len(v) > 0 {
                h.AddNode(v)
            }
        case int:
            if v > 0 {
                h.numSpots = v
            }
        case HashFunc:
            if v != nil {
                h.hashFunc = v
            }
        }
    }

    return h
}

type HashRing struct {
    numSpots int
    hashFunc HashFunc
    weights  map[string]int
    items    []hashItem
}

type hashItem struct {
    node  string
    value uint32
}

type HashFunc func(string) uint32

// add node to hash ring with default 1 weight,
// if node add multiple times, it gets
// a proportional amount of weight.
func (h *HashRing) AddNode(node string, w ...int) {
    weight := h.weights[node]
    if len(w) == 1 && w[0] > 0 {
        weight += w[0]
    } else {
        weight += 1
    }

    h.weights[node] = weight
    h.init()
}

// remove node from hash ring
func (h *HashRing) DelNode(node string) {
    delete(h.weights, node)
    h.init()
}

// get node from hash ring by the hash value of key
func (h *HashRing) GetNode(key string) string {
    if len(h.weights) == 0 {
        return ""
    }

    if len(h.weights) == 1 {
        return h.items[0].node
    }

    hl := len(h.items)
    hv := h.hashFunc(key)
    iv := sort.Search(hl, func(i int) bool {
        return h.items[i].value >= hv
    })

    return h.items[iv%hl].node
}

func (h *HashRing) init() {
    totalWeight := 0
    for _, w := range h.weights {
        totalWeight += w
    }

    totalSpots := h.numSpots * len(h.weights)
    h.items = h.items[:0]

    for n, w := range h.weights {
        spots := int(math.Round(float64(totalSpots) * float64(w) / float64(totalWeight)))
        if spots <= 0 {
            spots = 1
        }

        for i := 0; i < spots; i++ {
            v := h.hashFunc(fmt.Sprintf("%s:%d", n, i))
            h.items = append(h.items, hashItem{n, v})
        }
    }

    sort.Slice(h.items, func(i, j int) bool { return h.items[i].value < h.items[j].value })
}
