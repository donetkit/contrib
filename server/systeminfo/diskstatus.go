package systeminfo

type DiskStatus struct {
	All  uint64 `json:"all"`  // 总空间
	Used uint64 `json:"used"` // 可用空间
	Free uint64 `json:"free"` // 剩余空间
}
