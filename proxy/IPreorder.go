package proxy

import (
	"RFC9298proxy/utils"
	"encoding/binary"
)

// used to manage fragment ip packets.
type IPreorder struct {
	ReorderedPacket map[int][]byte
	FinishAssemble  bool
	Packet          []byte
}

//need to use copy, not assign
func (p *IPreorder) CheckFragment(buf []byte) bool {
	offset := calculateOffset(buf)
	id := parseId(buf)
	var flag uint8
	flag = buf[6] & 0x20
	var packetLen int
	if flag == 0x20 {
		utils.DebugPrintf("More Fragment bit is set!\n")
		// Assume packet receive in order
		if offset == 0 {
			IPpayload := buf
			p.ReorderedPacket[id] = make([]byte, len(IPpayload))
			copy(p.ReorderedPacket[id], IPpayload)
			packetLen = len(p.ReorderedPacket[id])
			utils.DebugPrintf("packet %d has %d payload.\n", id, packetLen)
		} else {
			_, ok := p.ReorderedPacket[id]
			if !ok {
				// The packet order is wrong, need to handle it.
				utils.ErrorPrintf("[Error] IP packet order is wrong.")
			}
			// may need other check
			IPpayload := buf[20:]
			p.ReorderedPacket[id] = append(p.ReorderedPacket[id], IPpayload...)
			packetLen = len(p.ReorderedPacket[id])
			utils.DebugPrintf("packet %d has %d payload.\n", id, packetLen)
		}
		return true
	}
	if offset != 0 {
		p.ReorderedPacket[id] = append(p.ReorderedPacket[id], buf[20:]...)
		p.Packet = make([]byte, len(p.ReorderedPacket[id]))
		copy(p.Packet, p.ReorderedPacket[id])
		p.FinishAssemble = true
		utils.DebugPrintf("total packet len:%d\n", len(p.Packet))
		// utils.DebugPrintf("packet payload:%x\n", p.Packet)
		return true
	}
	utils.DebugPrintf("More Fragment bit is not set!\n")
	return false
}

func calculateOffset(buf []byte) int {
	var offset int
	offsetByte := make([]byte, 2)
	offsetByte = buf[6:8]
	temp := offsetByte[0]
	offsetByte[0] = offsetByte[0] & 0x1F
	offset = int(binary.BigEndian.Uint16(offsetByte))
	offset *= 8
	buf[6] = temp
	utils.DebugPrintf("offset: %d\n", offset)
	return offset
}

func parseId(buf []byte) int {
	var id int
	id = int(binary.BigEndian.Uint16(buf[4:6]))
	utils.DebugPrintf("packet id: %d\n", id)
	return id
}
