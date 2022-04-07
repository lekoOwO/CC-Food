package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func initAPI(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	user := r.Group("/user")
	{
		user.GET("/", func(c *gin.Context) {
			users := []User{}
			db.Preload(clause.Associations).Preload("Purchases.PurchaseDetails").Find(&users, "is_disabled = ?", false)
			c.JSON(http.StatusOK, users)
		})

		user.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user := User{}
			db.Preload(clause.Associations).Preload("Purchases.PurchaseDetails").Take(&user, map[string]interface{}{
				"id":          id,
				"is_disabled": false,
			})
			c.JSON(http.StatusOK, user)
		})

		user.POST("/", func(c *gin.Context) {
			type request struct {
				Usernames   []string `json:"usernames"`
				DisplayName string   `json:"display_name"`
			}
			var data request
			c.BindJSON(&data)

			var user User
			var dbUserNames []Username
			for _, username := range data.Usernames {
				dbUserNames = append(dbUserNames, Username{
					Name: username,
				})
			}
			user = User{
				Usernames:   dbUserNames,
				DisplayName: data.DisplayName,
			}

			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, user)
		})

		user.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user := User{}

			if err := db.First(&user, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			if err := db.Model(&user).Updates(map[string]interface{}{
				"is_disabled": true,
			}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		user.GET("/byUsername/:username", func(c *gin.Context) {
			username := Username{}
			if err := db.Preload("User").Take(&username, "name = ?", c.Param("username")).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, username.User)
		})
	}

	username := r.Group("/username")
	{
		username.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			username := Username{}
			if err := db.Preload(clause.Associations).Take(&username, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, username)
		})

		username.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			db.Delete(&Username{}, id)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		username.POST("/:id", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user := User{}
			if err := db.Take(&user, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			json := make(map[string]interface{})
			c.BindJSON(&json)

			name := json["name"].(string)

			username := Username{
				Name: name,
			}
			db.Model(&user).Association("Usernames").Append(&username)

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	product := r.Group("/product")
	{
		product.GET("/", func(c *gin.Context) {
			products := []Product{}
			db.Find(&products, "is_disabled = ?", false)
			c.JSON(http.StatusOK, products)
		})

		product.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			product := Product{}
			if err := db.Take(&product, map[string]interface{}{
				"id":          id,
				"is_disabled": false,
			}).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, product)
		})

		product.POST("/", func(c *gin.Context) {
			product := Product{}
			c.BindJSON(&product)

			setPrivateBarcode := false

			if product.Barcode == "" {
				setPrivateBarcode = true
				product.Barcode = uuid.New().String()
			}

			db.Create(&product)

			if setPrivateBarcode {
				if err := db.Model(&product).Updates(map[string]interface{}{
					"barcode": fmt.Sprintf("%s%d", "CC-Food-", product.ID),
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
			c.JSON(http.StatusOK, product)
		})

		product.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			product := Product{}
			if err := db.Take(&product, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			if err := db.Model(&product).Updates(map[string]interface{}{
				"is_disabled": true,
			}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		product.PUT("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			product := Product{}
			if err := db.Take(&product, map[string]interface{}{
				"id":          id,
				"is_disabled": false,
			}).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			updateProduct := Product{}
			c.BindJSON(&updateProduct)

			if err := db.Model(&product).Updates(map[string]interface{}{
				"name":    updateProduct.Name,
				"price":   updateProduct.Price,
				"barcode": updateProduct.Barcode,
			}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		product.GET("/byBarcode/:barcode", func(c *gin.Context) {
			barcode := c.Param("barcode")
			product := Product{}
			if err := db.Take(&product, map[string]interface{}{
				"barcode":     barcode,
				"is_disabled": false,
			}).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, product)
		})
	}

	purchase := r.Group("/purchase")
	{
		purchase.GET("/", func(c *gin.Context) {
			purchases := []Purchase{}
			db.Preload("PurchaseDetails").Find(&purchases)
			c.JSON(http.StatusOK, purchases)
		})

		purchase.GET("/notPaid", func(c *gin.Context) {
			var response struct {
				Purchases []Purchase `json:"purchases"`
			}
			db.Preload("PurchaseDetails").Find(&response.Purchases, "payment_id IS NULL")
			c.JSON(http.StatusOK, response)
		})

		purchase.GET("/notPaid/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			purchases := []Purchase{}
			db.Preload("PurchaseDetails").Find(&purchases, "user_id = ? AND payment_id IS NULL", id)
			c.JSON(http.StatusOK, purchases)
		})

		purchase.GET("/id/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			purchase := Purchase{}
			if err := db.Preload("PurchaseDetails").Take(&purchase, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, purchase)
		})

		purchase.POST("/", func(c *gin.Context) {
			data := BuyRequest{}
			c.BindJSON(&data)

			user := User{}
			if err := db.First(&user, map[string]interface{}{
				"id":          data.UserID,
				"is_disabled": false,
			}).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			var pds []PurchaseDetail
			for _, brd := range data.Details {
				product := Product{}
				if err := db.First(&product, map[string]interface{}{
					"id":          brd.ProductID,
					"is_disabled": false,
				}).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
					return
				}

				pd := PurchaseDetail{
					Quantity: brd.Quantity,
					Total:    product.Price * brd.Quantity,
					Product:  product,
				}
				pds = append(pds, pd)
			}

			purchase := Purchase{
				User:            user,
				PurchaseDetails: pds,
			}
			db.Create(&purchase)
			c.JSON(http.StatusOK, purchase)
		})

		purchase.POST("/pay", func(c *gin.Context) {
			data := struct {
				UserID      uint64   `json:"user_id"`
				PurchaseIDs []uint64 `json:"purchase_ids"`
			}{}
			c.BindJSON(&data)

			user := User{}
			if err := db.Take(&user, data.UserID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			payment := Payment{
				User:      user,
				Purchases: []Purchase{},
			}
			for _, id := range data.PurchaseIDs {
				purchase := Purchase{}
				if err := db.Take(&purchase, "id = ? AND payment_id IS NULL", id).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
					return
				}
				payment.Purchases = append(payment.Purchases, purchase)
			}

			db.Create(&payment)
			c.JSON(http.StatusOK, payment)
		})
	}

	type Form struct {
		Files []*multipart.FileHeader `form:"files" binding:"required"`
	}

	r.POST("/import", func(c *gin.Context) {
		var form Form
		_ = c.ShouldBind(&form)

		var defaultProduct Product
		if err := db.Take(&defaultProduct, 1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, formFile := range form.Files {

			openedFile, _ := formFile.Open()
			file, _ := ioutil.ReadAll(openedFile)

			var oldSystemData OldSystemData
			json.Unmarshal(file, &oldSystemData)

			username := Username{}
			if err := db.Preload("User").Take(&username, "name = ?", oldSystemData.User).Error; err == nil {
				// username found, ignore.
				continue
			}

			var purchases []Purchase
			payments := map[float64]Payment{}
			for _, t := range oldSystemData.Transactions {
				var pd PurchaseDetail
				pd.Quantity = t.Amount
				pd.Total = t.Amount
				pd.Product = defaultProduct
				pd.CreatedAt = float64ToTime(t.CreatedAt)

				p := Purchase{
					PurchaseDetails: []PurchaseDetail{pd},
				}

				if t.DeletedAt != nil {
					deletedAt := *t.DeletedAt
					payment, ok := payments[deletedAt]
					if !ok {
						payment = Payment{}
						payment.CreatedAt = float64ToTime(deletedAt)
						payments[deletedAt] = payment
					}
					p.Payment = &payment
				}

				purchases = append(purchases, p)
			}

			var paymentKeys []float64
			var paymentValues []Payment
			for time := range payments {
				paymentKeys = append(paymentKeys, time)
			}

			sort.Float64s(paymentKeys)
			for _, time := range paymentKeys {
				paymentValues = append(paymentValues, payments[time])
			}

			// Update Payment reference
			for i := range purchases {
				p := &purchases[i]
				if p.Payment != nil {
					for j := range paymentValues {
						p2 := &paymentValues[j]
						if p2.CreatedAt.Equal(p.Payment.CreatedAt) {
							p.Payment = p2
							break
						}
					}
				}
			}

			user := User{
				Usernames: []Username{
					{
						Name: oldSystemData.User,
					},
				},
				DisplayName: oldSystemData.User,
				Purchases:   purchases,
				Payments:    paymentValues,
			}

			db.Create(&user)
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	return r
}
