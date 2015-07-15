package groupdb

type DayID uint32

func MakeDayID(year, month, day int) DayID {
	return (DayID(year)<<16)|(DayID(month)<<8)|DayID(day)
}

func (d DayID) Date() (year int, month int, day int) {
	year = int((d>>16)&0xffff)
	month = int((d>>8)&0xff)
	day = int(d&0xff)
	return
}

type DayProvider func() DayID

type Group struct{
	Name string
	Desc string
	Low int64
	High int64
	Count int64
	State byte
}

type GroupDB interface{
	AddGroup(group, descr string,state byte)
	Groups(prefix string,ptr *Group,cb func())
	Group(group string, ptr *Group) bool
	Numberate(groups []string, id string)
	GetArticleID(group string, num int64) string
	Erase(upto DayID)
}


