package conf

type Item struct {
	Id            int32  `json:"id"`
	Name          string `json:"name"`
	Info          string `json:"info"`
	Image         string `json:"image"`
	Rare          int32  `json:"rare"`
	Page          int32  `json:"page"`
	Type          int32  `json:"type"`
	UseValue      int32  `json:"useValue"`
	IsSell        bool   `json:"isSell"`
	SellValue     int32  `json:"sellValue"`
	IsComposite   bool   `json:"isComposite"`
	CompositeNeed int32  `json:"compositeNeed"`
	CompositeItem int32  `json:"compositeItem"`
	Show          bool   `json:"show"`
	UseJump       string `json:"useJump"`
}

// ItemE 通用物品类型
type ItemE struct {
	ItemId   int32
	ItemNum  int32
	ItemType int32
}

// 兑换码
type Cdkey struct {
	Id        int32    `json:"id"`
	Keys      []string `json:"keys"`
	Iscommon  bool     `json:"iscommon"`
	Count     int32    `json:"count"` // 兑换码总次数限制，0表示不限次
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
	Awards    []ItemE  `json:"awards"`
}
