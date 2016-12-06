package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"

	"../../cnutil"
)

type BlockTemplate struct {
	Difficulty     int64
	Height         int64
	ReservedOffset int
	PrevHash       string
	Buffer         []byte
}

func (b *BlockTemplate) nextBlob(extraNonce uint32, instanceId []byte) string {
	extraBuff := new(bytes.Buffer)
	binary.Write(extraBuff, binary.BigEndian, extraNonce)

	blobBuff := make([]byte, len(b.Buffer))
	copy(blobBuff, b.Buffer)
	copy(blobBuff[b.ReservedOffset+4:b.ReservedOffset+7], instanceId)
	copy(blobBuff[b.ReservedOffset:], extraBuff.Bytes())
	blob := cnutil.ConvertBlob(blobBuff)
	return hex.EncodeToString(blob)
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
		Difficulty:     reply.Difficulty,
		Height:         reply.Height,
		PrevHash:       reply.PrevHash,
		ReservedOffset: reply.ReservedOffset,
	}
	newTemplate.Buffer, _ = hex.DecodeString(reply.Blob)
	s.blockTemplate.Store(&newTemplate)
	return true
}
