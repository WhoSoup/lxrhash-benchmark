package main

// copied from pegnet repo to prevent having to import pegnet
type ninc struct {
	Nonce []byte
}

func newninc(id int) *ninc {
	n := new(ninc)
	n.Nonce = []byte{byte(id), 0}
	return n
}

func (i *ninc) next() {
	idx := len(i.Nonce) - 1
	for {
		i.Nonce[idx]++
		if i.Nonce[idx] == 0 {
			idx--
			if idx == 0 {
				rest := append([]byte{1}, i.Nonce[1:]...)
				i.Nonce = append([]byte{i.Nonce[0]}, rest...)
				break
			}
		} else {
			break
		}
	}

}
