package main

import "sync"

// Type of Packet Based on BuildTCPPacket.go
type pPacket record

// Struct for the Queue itself
type PacketQueue struct {
	queue []pPacket
	lock  sync.Mutex
}

// Create a new Queue object
func (q *PacketQueue) New() *PacketQueue {
	q.queue = []pPacket{}
	return q
}

// Append a new packet to the end of the Queue
func (q *PacketQueue) Append(pack pPacket) {
	q.lock.Lock()
	q.queue = append(q.queue, pack)
	q.lock.Unlock()
}

// Pops first packet added from the queue
func (q *PacketQueue) Pop() *pPacket {
	q.lock.Lock()
	packet := q.queue[0]
	q.queue = q.queue[1:len(q.queue)]
	q.lock.Unlock()

	return &packet
}

// Checks if the queue is empty
func (q *PacketQueue) IsEmpty() bool {
	return len(q.queue) == 0
}

// Return the Size of the Queue
func (q *PacketQueue) Size() int {
	return len(q.queue)
}
