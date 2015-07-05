package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"sync/atomic"

	"../../cnutil"
)

type BlockTemplate struct {
	Blob           string
	Difficulty     int64
	Height         int64
	ReservedOffset int
	PrevHash       string
	Buffer         []byte
	ExtraNonce     uint32
}

func (b *BlockTemplate) nextBlob() (string, uint32) {
	// Preventing race by using atomic op here
	// No need for using locks, because this is only one write to BT and it's atomic
	extraNonce := atomic.AddUint32(&b.ExtraNonce, 1)
	extraBuff := new(bytes.Buffer)
	binary.Write(extraBuff, binary.BigEndian, extraNonce)
	blobBuff := make([]byte, len(b.Buffer))
	copy(blobBuff, b.Buffer) // We never write to this buffer to prevent race
	copy(blobBuff[b.ReservedOffset:], extraBuff.Bytes())
	blob := cnutil.ConvertBlob(blobBuff)
	return hex.EncodeToString(blob), extraNonce
}

func (s *StratumServer) fetchBlockTemplate() bool {
	reply, err := s.rpc.GetBlockTemplate(8, s.config.Address)
	if err != nil {
		log.Printf("Error while refreshing block template: %s", err)
		return false
	}
	t := s.currentBlockTemplate()

	if t.PrevHash == reply.PrevHash {
		// Fallback to height comparison
		if len(reply.PrevHash) == 0 && reply.Height > t.Height {
			log.Printf("New block to mine at height %v, diff: %v", reply.Height, reply.Difficulty)
		} else {
			return false
		}
	} else {
		log.Printf("New block to mine at height %v, diff: %v, prev_hash: %s", reply.Height, reply.Difficulty, reply.PrevHash)
	}
	newTemplate := BlockTemplate{
		Blob:           reply.Blob,
		Difficulty:     reply.Difficulty,
		Height:         reply.Height,
		PrevHash:       reply.PrevHash,
		ReservedOffset: reply.ReservedOffset,
	}
	newTemplate.Buffer, _ = hex.DecodeString(reply.Blob)
	copy(newTemplate.Buffer[reply.ReservedOffset+4:reply.ReservedOffset+7], s.instanceId)
	newTemplate.ExtraNonce = 0
	s.blockTemplate.Store(&newTemplate)
	return true
}
