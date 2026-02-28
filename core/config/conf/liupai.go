package conf

type Liupai struct {
	Id          int32 `json:"Id"`
	Number      int32 `json:"Number"`
	JobAdd_Yao  int32 `json:"JobAdd_Yao"`
	JobAdd_Shen int32 `json:"JobAdd_Shen"`
	JobAdd_Fo   int32 `json:"JobAdd_Fo"`
}

type LiupaiRestrain struct {
	Id                 int32   `json:"Id"`
	Job                int32   `json:"Job"`
	Restrain           int32   `json:"Restrain"`
	Restrain_Attribute []int32 `json:"Restrain_Attribute"`
}
