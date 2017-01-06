package gopack

import (
	"container/heap"
	"sync"
	"time"
)

// newMemoryStorage creates and initializes a new memoryStorage
func newMemoryStorage() *memoryStorage {
	ms := new(memoryStorage)
	ms.index = make(map[int]int)
	ms.packets = make(map[int][]byte)
	return ms
}

// memoryStorage is used to save packet data
type memoryStorage struct {
	uniqueID int // incoming packet id
	index    map[int]int
	packets  map[int][]byte

	// A PriorityQueue implements heap.
	priorityQueue []*Packet

	muxUniqueID      sync.Mutex
	muxPriorityQueue sync.Mutex
	muxPackets       sync.Mutex
}

// Len is the number of elements in the priority queue
func (ms *memoryStorage) Len() int {
	return len(ms.priorityQueue)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ms *memoryStorage) Less(i, j int) bool {
	item1 := ms.priorityQueue[i]
	item2 := ms.priorityQueue[j]
	if item1.Confirm && !item2.Confirm {
		return true
	} else if !item1.Confirm && item2.Confirm {
		return false
	}
	if item1.Timestamp == item2.Timestamp {
		return item1.MsgID < item2.MsgID
	}
	return item1.Timestamp < item2.Timestamp
}

// Swap swaps the elements with indexes i and j
func (ms *memoryStorage) Swap(i, j int) {
	n := len(ms.priorityQueue)
	if i >= 0 && i < n && j >= 0 && j < n {
		ms.priorityQueue[i], ms.priorityQueue[j] = ms.priorityQueue[j], ms.priorityQueue[i]
		ms.index[ms.priorityQueue[i].MsgID] = i
		ms.index[ms.priorityQueue[j].MsgID] = j
	}
}

// Push add x as element Len()
func (ms *memoryStorage) Push(x interface{}) {
	index := len(ms.priorityQueue)
	packet := x.(*Packet)
	ms.priorityQueue = append(ms.priorityQueue, packet)
	ms.index[packet.MsgID] = index
}

// Pop remove and return element Len() - 1
func (ms *memoryStorage) Pop() interface{} {
	old := ms.priorityQueue
	n := len(old)
	if n > 0 {
		packet := old[n-1]
		ms.priorityQueue = old[0 : n-1]
		return packet
	}
	return nil
}

// UniqueID generate unique id for new packet
func (ms *memoryStorage) UniqueID() int {
	ms.muxUniqueID.Lock()
	defer ms.muxUniqueID.Unlock()
	ms.uniqueID++
	return ms.uniqueID
}

// Save insert packet into queue
func (ms *memoryStorage) Save(packet *Packet) {
	ms.muxPriorityQueue.Lock()
	defer ms.muxPriorityQueue.Unlock()
	heap.Push(ms, packet)
}

// Unconfirmed is used to return latest unconfirmed packet
func (ms *memoryStorage) Unconfirmed() *Packet {
	ms.muxPriorityQueue.Lock()
	defer ms.muxPriorityQueue.Unlock()
	for {
		packet, ok := heap.Pop(ms).(*Packet)
		if ok {
			delete(ms.index, packet.MsgID)
			if packet.Confirm {
				continue
			} else {
				if packet.Timestamp > time.Now().Unix() {
					heap.Push(ms, packet)
				} else {
					return packet
				}
			}
		}
		return nil
	}
}

// Confirm is used to set element.Confirm to true and Fix priority queue
func (ms *memoryStorage) Confirm(id int) *Packet {
	ms.muxPriorityQueue.Lock()
	defer ms.muxPriorityQueue.Unlock()
	index, ok := ms.index[id]
	if ok {
		packet := ms.priorityQueue[index]
		packet.Confirm = true
		heap.Fix(ms, index)
		return packet
	}
	return nil
}

// Receive and save packet
func (ms *memoryStorage) Receive(id int, payload []byte) {
	ms.muxPackets.Lock()
	defer ms.muxPackets.Unlock()
	ms.packets[id] = payload
}

// Release and delete packet
func (ms *memoryStorage) Release(id int) []byte {
	ms.muxPackets.Lock()
	defer ms.muxPackets.Unlock()
	packet := ms.packets[id]
	delete(ms.packets, id)
	return packet
}
