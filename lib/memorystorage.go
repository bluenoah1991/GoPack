package gopack

import "sync"
import "container/heap"

// MemoryStorage is used to save packet data
type MemoryStorage struct {
	uniqueID uint16 // incoming packet id
	index    map[uint16]int
	packets  map[uint16][]byte

	// A PriorityQueue implements heap.
	priorityQueue []*Packet

	muxUniqueID      sync.Mutex
	muxPriorityQueue sync.Mutex
	muxPackets       sync.Mutex
}

// Len is the number of elements in the priority queue
func (ms *MemoryStorage) Len() int {
	return len(ms.priorityQueue)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ms *MemoryStorage) Less(i, j int) bool {
	item1 := ms.priorityQueue[i]
	item2 := ms.priorityQueue[j]
	if item1.Confirm && !item2.Confirm {
		return true
	} else if !item1.Confirm && item2.Confirm {
		return false
	}
	if item1.Timestamp == item2.Timestamp {
		return item1.MsgID > item2.MsgID
	}
	return item1.Timestamp > item2.Timestamp
}

// Swap swaps the elements with indexes i and j
func (ms *MemoryStorage) Swap(i, j int) {
	ms.priorityQueue[i], ms.priorityQueue[j] = ms.priorityQueue[j], ms.priorityQueue[i]
	ms.index[ms.priorityQueue[i].MsgID] = i
	ms.index[ms.priorityQueue[j].MsgID] = j
}

// Push add x as element Len()
func (ms *MemoryStorage) Push(x interface{}) {
	index := len(ms.priorityQueue)
	packet := x.(*Packet)
	ms.priorityQueue = append(ms.priorityQueue, packet)
	ms.index[packet.MsgID] = index
}

// Pop remove and return element Len() - 1
func (ms *MemoryStorage) Pop() interface{} {
	old := ms.priorityQueue
	n := len(old)
	packet := old[n-1]
	ms.priorityQueue = old[0 : n-1]
	return packet
}

// UniqueID generate unique id for new packet
func (ms *MemoryStorage) UniqueID() uint16 {
	ms.muxUniqueID.Lock()
	defer ms.muxUniqueID.Unlock()
	ms.uniqueID++
	return ms.uniqueID
}

// Save insert packet into queue
func (ms *MemoryStorage) Save(packet *Packet) {
	ms.muxPriorityQueue.Lock()
	defer ms.muxPriorityQueue.Unlock()
	heap.Push(ms, packet)
}

// Unconfirmed is used to return latest unconfirmed packet
func (ms *MemoryStorage) Unconfirmed() *Packet {
	ms.muxPriorityQueue.Lock()
	defer ms.muxPriorityQueue.Unlock()
	for {
		packet := heap.Pop(ms).(*Packet)
		if packet == nil {
			return nil
		} else if packet.Confirm {
			continue
		} else {
			return packet
		}
	}
}

// Confirm is used to set element.Confirm to true and Fix priority queue
func (ms *MemoryStorage) Confirm(id uint16) *Packet {
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
func (ms *MemoryStorage) Receive(id uint16, payload []byte) {
	ms.muxPackets.Lock()
	defer ms.muxPackets.Unlock()
	ms.packets[id] = payload
}

// Release and delete packet
func (ms *MemoryStorage) Release(id uint16) []byte {
	ms.muxPackets.Lock()
	defer ms.muxPackets.Unlock()
	packet := ms.packets[id]
	delete(ms.packets, id)
	return packet
}
