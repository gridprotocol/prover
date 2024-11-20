package prover

/*
int generatePOW(char *rand, int len, int diffcult, long long *index);
#cgo LDFLAGS: -L../.. -lpow
*/
import "C"
import (
	"grid-prover/core/types"
	"unsafe"

	"golang.org/x/xerrors"
)

func GeneratePOW(nodeID types.NodeID, rand []byte, diffcult int) (int64, error) {
	var res int64
	var prefixBuf = append(rand, nodeID.ToBytes()...)
	c_prefix := (*C.char)(unsafe.Pointer(&prefixBuf[0]))
	c_index := (*C.longlong)(unsafe.Pointer(&res))
	if C.generatePOW(c_prefix, C.int(len(prefixBuf)), C.int(diffcult), c_index) != 0 {
		return 0, xerrors.New("Unexpected Error")
	}
	return res, nil
}
