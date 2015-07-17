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
	Name  string
	Desc  string
	Low   int64
	High  int64
	Count int64
	State byte
}

type GroupMeta struct{
	Name  string
	Desc  string
	State byte
}

type GroupDB interface{
	// Adds a stream of groups, supplied using a channel. The stream MAY
	// supply groups, that do exist already in the Database, as AddGroups
	// is required to check each existing group, supplied through the channel.
	AddGroups(src <- chan GroupMeta)
	// Adds a group to the DB. The group MUST not exist already in the
	// Database as AddGroup is not required to check the existence of
	// the newsgroup.
	AddGroup(group, descr string,state byte)
	// Scans for all groups, that have the given prefix. 
	// For every group, that has been read, the function cb is called.
	Groups(prefix string,ptr *Group,cb func())
	// Reads a single group and stores it in the supplied Group object.
	// Returns true if the group exists.
	Group(group string, ptr *Group) bool
	Numberate(groups []string, id string)
	GetArticleID(group string, num int64) string
	Erase(upto DayID)
}


