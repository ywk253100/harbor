package models

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(
		new(Registry),
		new(RepPolicy),
		new(ScheduleJob))
}
