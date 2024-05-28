package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

type ArticleStatus uint8

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

const (
	ArticleStatusUnknown     = iota // 这是一个未知状态
	ArticleStatusUnpublished        // 未发表
	ArticleStatusPublished          // 已发表
	ArticleStatusPrivate            // 仅自己可见

)

type Author struct {
	Id   int64
	Name string
}
