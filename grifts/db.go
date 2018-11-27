package grifts

import (
	"fmt"

	"github.com/markbates/grift/grift"
	"github.com/monarko/piia/models"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		return nil
	})

})

var _ = grift.Namespace("db", func() {

	grift.Desc("clean", "Clean the database")
	grift.Add("clean", func(c *grift.Context) error {
		err := models.DB.TruncateAll()
		if err != nil {
			fmt.Println("=> Cleaning failed: ", err)
		}
		return nil
	})

})
