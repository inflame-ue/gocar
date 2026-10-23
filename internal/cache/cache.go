package cache

type L1 struct {
	root string
}

func NewL1(root string) *L1 {
	return &L1{
		root: root,
	}
}

