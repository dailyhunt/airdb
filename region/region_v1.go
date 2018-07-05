package region

type regionV1 struct {
	seqId int
}

func (r *regionV1) Close() {
	panic("implement me")
}

func (r *regionV1) Drop() {
	panic("implement me")
}

func (r *regionV1) Archive() {
	panic("implement me")
}

func (r *regionV1) Mutate() {
	panic("implement me")
}

func (r *regionV1) Get() {
	panic("implement me")
}

func (r *regionV1) Merge() {
	panic("implement me")
}
