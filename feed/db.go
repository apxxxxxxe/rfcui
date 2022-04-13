package feed

import (
	"bytes"
	"encoding/gob"
)

func EncodeFeed(feeds *Feed) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(feeds)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeFeed(data []byte) *Feed {
	var feeds Feed
	buf := bytes.NewBuffer(data)
	_ = gob.NewDecoder(buf).Decode(&feeds)
	return &feeds
}

func EncodeGroup(groups *Group) ([]byte, error) {
  buf := bytes.NewBuffer(nil)
  enc := gob.NewEncoder(buf)
  err := enc.Encode(groups)
  if err != nil {
    return nil, err
  }
  return buf.Bytes(), nil
}

func DecodeGroup(data []byte) *Group {
  var groups Group
  buf := bytes.NewBuffer(data)
  _ = gob.NewDecoder(buf).Decode(&groups)
  return &groups
}
