package services

import (
	"github.com/fpay/gopress"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	. "go_bbs/conf"
)

const (
	// DatabaseServiceName is the identity of database service
	DatabaseServiceName = "database"
)

// DatabaseService type
type DatabaseService struct {
	// Uncomment this line if this service has dependence on other services in the container
	// c *gopress.Container

	DB *gorm.DB
}

// NewDatabaseService returns instance of database service
func NewDatabaseService() *DatabaseService {
	var err error
	ds := new(DatabaseService)
	sqlConnection := Conf.DB.UserName + ":" + Conf.DB.Pwd + "@tcp(" + Conf.DB.Host + ":" + Conf.DB.Port + ")/" + Conf.DB.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
	ds.DB, err = gorm.Open("mysql", sqlConnection)
	//ds.DB, err = gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/gobbs?charset=utf8&parseTime=True&loc=Local")
	//defer ds.DBS.Close()
	if err != nil {
		panic(err)
	}
	return ds
}

// ServiceName is used to implements gopress.Service
func (s *DatabaseService) ServiceName() string {
	return DatabaseServiceName
}

// RegisterContainer is used to implements gopress.Service
func (s *DatabaseService) RegisterContainer(c *gopress.Container) {
	// Uncomment this line if this service has dependence on other services in the container
	// s.c = c
}

func (s *DatabaseService) SampleMethod() {
}
