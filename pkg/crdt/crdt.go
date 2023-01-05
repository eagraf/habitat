package crdt

// #cgo CFLAGS: -I${SRCDIR}
// #cgo LDFLAGS: -L${SRCDIR} -lyrs_osx -lyrs_linux
// #include "libyrs.h"
import "C"
import (
	"fmt"
	"unsafe"
)

type Doc struct {
	ydoc *C.struct_YDoc
}

func NewDoc(state []byte) *Doc {
	doc := &Doc{ydoc: C.ydoc_new()}
	doc.Merge(state)
	return doc
}

func (doc *Doc) Merge(state []byte) {
	if len(state) == 0 {
		return
	}
	txn := C.ydoc_write_transaction(doc.ydoc)
	err := C.ytransaction_apply_v2(txn, (*C.uchar)(unsafe.Pointer(&state[0])), C.int(len(state)))
	if int(err) > 0 {
		fmt.Println("Transaction apply failed with err code: ", err)
	}
	C.ytransaction_commit(txn)
}

func (doc *Doc) State() []byte {
	txn := C.ydoc_read_transaction(doc.ydoc)
	defer C.ytransaction_commit(txn)
	var len C.int
	diff := C.ytransaction_state_diff_v2(txn, nil, C.int(0), (*C.int)(unsafe.Pointer(&len)))
	bytes := append([]byte(nil), C.GoBytes(unsafe.Pointer(diff), len)...)
	C.ybinary_destroy(diff, len)

	return bytes
}

func (doc *Doc) Close() {
	C.ydoc_destroy(doc.ydoc)
}

func (doc *Doc) GetTextValue(name string) (output string) {
	branch := C.ytext(doc.ydoc, C.CString(name))
	txn := C.ybranch_read_transaction(branch)
	defer C.ytransaction_commit(txn)

	value := C.ytext_string(branch, txn)
	output = C.GoString(value)
	C.ystring_destroy(value)

	return
}

// Returns start, divergence A, divergence B, merged
func TestStates() (start []byte, a []byte, b []byte, merged []byte) {
	docA := C.ydoc_new()
	defer C.ydoc_destroy(docA)

	branchA := C.ytext(docA, C.CString("text"))
	startTxn := C.ydoc_write_transaction(docA)
	C.ytext_insert(branchA, startTxn, C.int(0), C.CString("hlo"), nil)
	var len C.int
	startDiff := C.ytransaction_state_diff_v2(startTxn, nil, C.int(0), (*C.int)(unsafe.Pointer(&len)))
	start = append([]byte(nil), C.GoBytes(unsafe.Pointer(startDiff), len)...)
	defer C.ybinary_destroy(startDiff, len)
	C.ytransaction_commit(startTxn)

	aTxn := C.ydoc_write_transaction(docA)
	C.ytext_insert(branchA, aTxn, C.int(1), C.CString("el"), nil)
	aDiff := C.ytransaction_state_diff_v2(aTxn, nil, C.int(0), (*C.int)(unsafe.Pointer(&len)))
	a = append([]byte(nil), C.GoBytes(unsafe.Pointer(aDiff), len)...)
	C.ybinary_destroy(aDiff, len)
	C.ytransaction_commit(aTxn)

	docB := C.ydoc_new()
	defer C.ydoc_destroy(docB)
	docBTxn := C.ydoc_write_transaction(docB)
	err1 := C.ytransaction_apply_v2(docBTxn, startDiff, len)
	C.ytransaction_commit(docBTxn)
	if int(err1) > 0 {
		fmt.Println("tests states error", err1)
	}

	branchB := C.ytext(docB, C.CString("text"))
	bTxn := C.ydoc_write_transaction(docB)
	C.ytext_insert(branchB, bTxn, C.int(3), C.CString(" world"), nil)
	bDiff := C.ytransaction_state_diff_v2(bTxn, nil, C.int(0), (*C.int)(unsafe.Pointer(&len)))
	defer C.ybinary_destroy(bDiff, len)
	b = append([]byte(nil), C.GoBytes(unsafe.Pointer(bDiff), len)...)
	C.ytransaction_commit(bTxn)

	mergeTxn := C.ydoc_write_transaction(docA)
	err2 := C.ytransaction_apply_v2(mergeTxn, bDiff, len)
	if int(err2) > 0 {
		fmt.Println("tests states error", err2)
	}
	mergeDiff := C.ytransaction_state_diff_v2(mergeTxn, nil, C.int(0), (*C.int)(unsafe.Pointer(&len)))
	merged = append([]byte(nil), C.GoBytes(unsafe.Pointer(mergeDiff), len)...)
	C.ybinary_destroy(mergeDiff, len)
	C.ytransaction_commit(mergeTxn)

	return
}
